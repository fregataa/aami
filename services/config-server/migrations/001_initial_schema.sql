-- AAMI Config Server - Unified Database Schema
-- Consolidated from migrations 001-009
-- Created: 2024-12-29
-- Description: Complete schema with all tables, indexes, and default data

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- TABLES
-- ============================================================================

-- Namespaces table: Logical grouping of groups
CREATE TABLE IF NOT EXISTS namespaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    policy_priority INTEGER NOT NULL DEFAULT 100,
    merge_strategy VARCHAR(20) NOT NULL DEFAULT 'merge' CHECK (merge_strategy IN ('override', 'merge')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Groups table: Hierarchical organization with namespace
CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    namespace_id UUID REFERENCES namespaces(id) ON DELETE RESTRICT,
    parent_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    description TEXT,
    priority INTEGER NOT NULL DEFAULT 100,
    is_default_own BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Targets table: Monitored servers/nodes
CREATE TABLE IF NOT EXISTS targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hostname VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(45) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'down')),
    labels JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    last_seen TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Target-Group junction table (many-to-many)
CREATE TABLE IF NOT EXISTS target_groups (
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    is_default_own BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (target_id, group_id)
);

-- Exporters table: Metric collectors configuration
CREATE TABLE IF NOT EXISTS exporters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('node_exporter', 'dcgm_exporter', 'custom')),
    port INTEGER NOT NULL CHECK (port > 0 AND port <= 65535),
    enabled BOOLEAN NOT NULL DEFAULT true,
    metrics_path VARCHAR(255) NOT NULL DEFAULT '/metrics',
    scrape_interval VARCHAR(20) NOT NULL DEFAULT '15s',
    scrape_timeout VARCHAR(20) NOT NULL DEFAULT '10s',
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Alert templates table: Reusable alert definitions
CREATE TABLE IF NOT EXISTS alert_templates (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('critical', 'warning', 'info')),
    query_template TEXT NOT NULL,
    default_config JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Alert rules table: Group-specific alert configurations (with template snapshot)
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    template_id VARCHAR(100) REFERENCES alert_templates(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('critical', 'warning', 'info')),
    query_template TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB NOT NULL DEFAULT '{}',
    merge_strategy VARCHAR(20) NOT NULL DEFAULT 'override' CHECK (merge_strategy IN ('override', 'merge')),
    priority INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Check templates table: Reusable check script definitions
CREATE TABLE IF NOT EXISTS check_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    check_type VARCHAR(255) NOT NULL,
    script_content TEXT NOT NULL,
    language VARCHAR(50) NOT NULL DEFAULT 'bash',
    default_config JSONB NOT NULL DEFAULT '{}',
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Check instances table: Check configurations at specific scopes (with template snapshot)
CREATE TABLE IF NOT EXISTS check_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID REFERENCES check_templates(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    check_type VARCHAR(100) NOT NULL,
    script_content TEXT NOT NULL,
    language VARCHAR(50) NOT NULL DEFAULT 'bash',
    scope VARCHAR(20) NOT NULL CHECK (scope IN ('global', 'namespace', 'group')),
    namespace_id UUID REFERENCES namespaces(id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    config JSONB NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    merge_strategy VARCHAR(20) NOT NULL DEFAULT 'merge' CHECK (merge_strategy IN ('override', 'merge')),
    priority INTEGER NOT NULL DEFAULT 100,
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Bootstrap tokens table: Auto-registration tokens
CREATE TABLE IF NOT EXISTS bootstrap_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    default_group_id UUID NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    max_uses INTEGER NOT NULL CHECK (max_uses > 0),
    uses INTEGER NOT NULL DEFAULT 0 CHECK (uses >= 0),
    expires_at TIMESTAMP NOT NULL,
    labels JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Namespaces indexes
CREATE INDEX IF NOT EXISTS idx_namespaces_name ON namespaces(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_namespaces_deleted_at ON namespaces(deleted_at);

-- Groups indexes
CREATE INDEX IF NOT EXISTS idx_groups_namespace ON groups(namespace_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_groups_parent ON groups(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_groups_name ON groups(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_groups_is_default_own ON groups(is_default_own) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_groups_deleted_at ON groups(deleted_at);

-- Targets indexes
CREATE INDEX IF NOT EXISTS idx_targets_hostname ON targets(hostname) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_targets_status ON targets(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_targets_deleted_at ON targets(deleted_at);

-- Target-Groups junction indexes
CREATE INDEX IF NOT EXISTS idx_target_groups_target ON target_groups(target_id);
CREATE INDEX IF NOT EXISTS idx_target_groups_group ON target_groups(group_id);

-- Exporters indexes
CREATE INDEX IF NOT EXISTS idx_exporters_target ON exporters(target_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_exporters_type ON exporters(type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_exporters_deleted_at ON exporters(deleted_at);

-- Alert templates indexes
CREATE INDEX IF NOT EXISTS idx_alert_templates_name ON alert_templates(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_alert_templates_deleted_at ON alert_templates(deleted_at);

-- Alert rules indexes
CREATE INDEX IF NOT EXISTS idx_alert_rules_group ON alert_rules(group_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_alert_rules_template ON alert_rules(template_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_alert_rules_deleted_at ON alert_rules(deleted_at);

-- Check templates indexes
CREATE INDEX IF NOT EXISTS idx_check_templates_check_type ON check_templates(check_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_check_templates_name ON check_templates(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_check_templates_deleted_at ON check_templates(deleted_at);

-- Check instances indexes
CREATE INDEX IF NOT EXISTS idx_check_instances_template ON check_instances(template_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_check_instances_scope ON check_instances(scope) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_check_instances_namespace ON check_instances(namespace_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_check_instances_group ON check_instances(group_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_check_instances_deleted_at ON check_instances(deleted_at);

-- Bootstrap tokens indexes
CREATE INDEX IF NOT EXISTS idx_bootstrap_tokens_token ON bootstrap_tokens(token) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_bootstrap_tokens_expires ON bootstrap_tokens(expires_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_bootstrap_tokens_deleted_at ON bootstrap_tokens(deleted_at);

-- ============================================================================
-- DEFAULT DATA
-- ============================================================================

-- Insert default alert templates (idempotent with ON CONFLICT DO NOTHING)
INSERT INTO alert_templates (id, name, description, severity, query_template, default_config) VALUES
('high_cpu_usage', 'High CPU Usage', 'Alert when CPU usage exceeds threshold', 'warning',
 '100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > {{ .threshold }}',
 '{"threshold": 80, "for": "5m"}'::jsonb),

('high_memory_usage', 'High Memory Usage', 'Alert when memory usage exceeds threshold', 'warning',
 '(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > {{ .threshold }}',
 '{"threshold": 90, "for": "5m"}'::jsonb),

('low_disk_space', 'Low Disk Space', 'Alert when disk space is running low', 'warning',
 '(node_filesystem_avail_bytes{fstype!~"tmpfs|fuse.lxcfs|squashfs|devtmpfs"} / node_filesystem_size_bytes) * 100 < {{ .threshold }}',
 '{"threshold": 20, "for": "5m"}'::jsonb),

('node_down', 'Node Down', 'Alert when node is unreachable', 'critical',
 'up{job="node-exporter"} == 0',
 '{"for": "2m"}'::jsonb),

('high_gpu_temperature', 'High GPU Temperature', 'Alert when GPU temperature exceeds threshold', 'critical',
 'dcgm_gpu_temp > {{ .threshold }}',
 '{"threshold": 85, "for": "3m"}'::jsonb),

('high_gpu_memory', 'High GPU Memory Usage', 'Alert when GPU memory usage exceeds threshold', 'warning',
 '(dcgm_fb_used / dcgm_fb_total) * 100 > {{ .threshold }}',
 '{"threshold": 90, "for": "5m"}'::jsonb)
ON CONFLICT (id) DO NOTHING;
