package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/fregataa/aami/config-server/internal/domain"
)

// NewTestNamespace creates a test namespace with default values
func NewTestNamespace(name string, policyPriority int) *domain.Namespace {
	return &domain.Namespace{
		ID:             uuid.New().String(),
		Name:           name,
		Description:    "Test namespace: " + name,
		PolicyPriority: policyPriority,
		MergeStrategy:  domain.MergeStrategyMerge,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// NewTestGroup creates a test group with default values
func NewTestGroup(name string, namespaceID string) *domain.Group {
	return &domain.Group{
		ID:           uuid.New().String(),
		Name:         name,
		NamespaceID:  namespaceID,
		Description:  "Test group: " + name,
		Priority:     100,
		IsDefaultOwn: false,
		Metadata:     make(map[string]string),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// NewTestGroupWithParent creates a test group with a parent
func NewTestGroupWithParent(name string, namespaceID string, parentID string) *domain.Group {
	group := NewTestGroup(name, namespaceID)
	group.ParentID = &parentID
	return group
}

// NewTestTarget creates a test target with default values
func NewTestTarget(hostname string, ipAddress string, groups []domain.Group) *domain.Target {
	return &domain.Target{
		ID:        uuid.New().String(),
		Hostname:  hostname,
		IPAddress: ipAddress,
		Groups:    groups,
		Status:    domain.TargetStatusActive,
		Labels:    make(map[string]string),
		Metadata:  make(map[string]string),
		LastSeen:  nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewTestTargetWithDefaultGroup creates a test target with a default own group
func NewTestTargetWithDefaultGroup(hostname string, ipAddress string, namespace *domain.Namespace) *domain.Target {
	defaultGroup := &domain.Group{
		ID:           uuid.New().String(),
		Name:         "target-" + hostname,
		NamespaceID:  namespace.ID,
		Description:  "Default group for " + hostname,
		Priority:     100,
		IsDefaultOwn: true,
		Metadata:     make(map[string]string),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return &domain.Target{
		ID:        uuid.New().String(),
		Hostname:  hostname,
		IPAddress: ipAddress,
		Groups:    []domain.Group{*defaultGroup},
		Status:    domain.TargetStatusActive,
		Labels:    make(map[string]string),
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewTestExporter creates a test exporter with default values
func NewTestExporter(targetID string, exporterType domain.ExporterType, port int) *domain.Exporter {
	return &domain.Exporter{
		ID:             uuid.New().String(),
		TargetID:       targetID,
		Type:           exporterType,
		Port:           port,
		Enabled:        true,
		MetricsPath:    "/metrics",
		ScrapeInterval: "15s",
		ScrapeTimeout:  "10s",
		Config:         domain.ExporterConfig{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// NewTestAlertTemplate creates a test alert template
func NewTestAlertTemplate(id string, severity domain.AlertSeverity) *domain.AlertTemplate {
	return &domain.AlertTemplate{
		ID:            id,
		Name:          "Test Alert: " + id,
		Description:   "Test alert template",
		Severity:      severity,
		QueryTemplate: "test_metric > {{ .threshold }}",
		DefaultConfig: domain.AlertRuleConfig{
			ForDuration: "5m",
			TemplateVars: map[string]interface{}{
				"threshold": 80,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewTestAlertRule creates a test alert rule
func NewTestAlertRule(groupID string, templateID string) *domain.AlertRule {
	return &domain.AlertRule{
		ID:      uuid.New().String(),
		GroupID: groupID,

		// Template fields (copied at creation)
		Name:          "Test Alert",
		Description:   "Test alert rule",
		Severity:      domain.AlertSeverityCritical,
		QueryTemplate: "SELECT * FROM metrics WHERE value > {{.threshold}}",
		DefaultConfig: domain.AlertRuleConfig{
			TemplateVars: map[string]interface{}{
				"threshold": 80,
			},
		},

		// Rule-specific fields
		Enabled: true,
		Config: domain.AlertRuleConfig{
			TemplateVars: map[string]interface{}{
				"threshold": 90,
			},
		},
		MergeStrategy: "override",
		Priority:      100,

		// Metadata
		CreatedFromTemplateID:   &templateID,
		CreatedFromTemplateName: strPtr("Test Template"),

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Helper function for string pointer
func strPtr(s string) *string {
	return &s
}

// NewTestBootstrapToken creates a test bootstrap token
func NewTestBootstrapToken(name string) *domain.BootstrapToken {
	token, _ := domain.GenerateToken()
	return &domain.BootstrapToken{
		ID:        uuid.New().String(),
		Token:     token,
		Name:      name,
		MaxUses:   10,
		Uses:      0,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Labels:    make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// TimePtr returns a pointer to a time
func TimePtr(t time.Time) *time.Time {
	return &t
}
