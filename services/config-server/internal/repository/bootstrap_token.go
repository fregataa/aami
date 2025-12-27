package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// BootstrapTokenModel is the GORM model for database operations
type BootstrapTokenModel struct {
	ID             string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Token          string         `gorm:"not null;uniqueIndex"`
	Name           string         `gorm:"not null"`
	DefaultGroupID string         `gorm:"type:uuid;not null;index"`
	DefaultGroup   *GroupModel    `gorm:"foreignKey:DefaultGroupID"`
	MaxUses        int            `gorm:"not null"`
	Uses           int            `gorm:"not null;default:0"`
	ExpiresAt      time.Time      `gorm:"not null;index"`
	Labels         StringMap      `gorm:"type:jsonb;default:'{}'"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (BootstrapTokenModel) TableName() string {
	return "bootstrap_tokens"
}

// ToBootstrapTokenModel converts domain.BootstrapToken to BootstrapTokenModel
func ToBootstrapTokenModel(bt *domain.BootstrapToken) *BootstrapTokenModel {
	model := &BootstrapTokenModel{
		ID:             bt.ID,
		Token:          bt.Token,
		Name:           bt.Name,
		DefaultGroupID: bt.DefaultGroupID,
		MaxUses:        bt.MaxUses,
		Uses:           bt.Uses,
		ExpiresAt:      bt.ExpiresAt,
		Labels:         StringMap(bt.Labels),
		CreatedAt:      bt.CreatedAt,
		UpdatedAt:      bt.UpdatedAt,
	}
	if bt.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *bt.DeletedAt, Valid: true}
	}
	return model
}

// ToDomain converts BootstrapTokenModel to domain.BootstrapToken
func (m *BootstrapTokenModel) ToDomain() *domain.BootstrapToken {
	bt := &domain.BootstrapToken{
		ID:             m.ID,
		Token:          m.Token,
		Name:           m.Name,
		DefaultGroupID: m.DefaultGroupID,
		MaxUses:        m.MaxUses,
		Uses:           m.Uses,
		ExpiresAt:      m.ExpiresAt,
		Labels:         map[string]string(m.Labels),
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		deletedAt := m.DeletedAt.Time
		bt.DeletedAt = &deletedAt
	}

	// Convert DefaultGroup if loaded
	if m.DefaultGroup != nil {
		bt.DefaultGroup = *m.DefaultGroup.ToDomain()
	}

	return bt
}

// BootstrapTokenRepository defines the interface for bootstrap token data access
type BootstrapTokenRepository interface {
	Create(ctx context.Context, token *domain.BootstrapToken) error
	GetByID(ctx context.Context, id string) (*domain.BootstrapToken, error)
	GetByToken(ctx context.Context, token string) (*domain.BootstrapToken, error)
	GetByGroupID(ctx context.Context, groupID string) ([]domain.BootstrapToken, error)
	Update(ctx context.Context, token *domain.BootstrapToken) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.BootstrapToken, int, error)
}

// bootstrapTokenRepository implements BootstrapTokenRepository interface using GORM
type bootstrapTokenRepository struct {
	db *gorm.DB
}

// NewBootstrapTokenRepository creates a new BootstrapTokenRepository instance
func NewBootstrapTokenRepository(db *gorm.DB) BootstrapTokenRepository {
	return &bootstrapTokenRepository{db: db}
}

func (r *bootstrapTokenRepository) Create(ctx context.Context, token *domain.BootstrapToken) error {
	model := ToBootstrapTokenModel(token)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	*token = *model.ToDomain()
	return nil
}

func (r *bootstrapTokenRepository) GetByID(ctx context.Context, id string) (*domain.BootstrapToken, error) {
	var model BootstrapTokenModel
	err := r.db.WithContext(ctx).
		Preload("DefaultGroup").
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *bootstrapTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.BootstrapToken, error) {
	var model BootstrapTokenModel
	err := r.db.WithContext(ctx).
		Preload("DefaultGroup").
		First(&model, "token = ?", tokenStr).Error
	if err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *bootstrapTokenRepository) GetByGroupID(ctx context.Context, groupID string) ([]domain.BootstrapToken, error) {
	var models []BootstrapTokenModel
	err := r.db.WithContext(ctx).
		Where("default_group_id = ?", groupID).
		Preload("DefaultGroup").
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	tokens := make([]domain.BootstrapToken, len(models))
	for i, model := range models {
		tokens[i] = *model.ToDomain()
	}
	return tokens, nil
}

func (r *bootstrapTokenRepository) Update(ctx context.Context, token *domain.BootstrapToken) error {
	model := ToBootstrapTokenModel(token)
	return r.db.WithContext(ctx).
		Model(&BootstrapTokenModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error
}

// Delete performs soft delete on a bootstrap token (sets deleted_at timestamp)
func (r *bootstrapTokenRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&BootstrapTokenModel{}, "id = ?", id).Error
}

// Purge permanently removes a bootstrap token from the database
func (r *bootstrapTokenRepository) Purge(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Delete(&BootstrapTokenModel{}, "id = ?", id).Error
}

// Restore restores a soft-deleted bootstrap token
func (r *bootstrapTokenRepository) Restore(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(&BootstrapTokenModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

func (r *bootstrapTokenRepository) List(ctx context.Context, page, limit int) ([]domain.BootstrapToken, int, error) {
	var models []BootstrapTokenModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&BootstrapTokenModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Preload("DefaultGroup").
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	tokens := make([]domain.BootstrapToken, len(models))
	for i, model := range models {
		tokens[i] = *model.ToDomain()
	}

	return tokens, int(total), nil
}
