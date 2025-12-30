-- AAMI Config Server - Unified Database Schema
-- Updated: 2025-12-30
-- Description: Complete schema with all tables, indexes, and default data
-- NOTE: This is the current production schema after namespace removal refactoring

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- TABLES
-- ============================================================================

-- Groups table: Flat organization (no hierarchy)
CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
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
    created_from_template_id VARCHAR(100) REFERENCES alert_templates(id) ON DELETE SET NULL,
    created_from_template_name VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('critical', 'warning', 'info')),
    query_template TEXT NOT NULL,
    default_config JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB NOT NULL DEFAULT '{}',
    merge_strategy VARCHAR(20) NOT NULL DEFAULT 'override' CHECK (merge_strategy IN ('override', 'merge')),
    priority INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Script templates table: Reusable script definitions (renamed from check_templates/monitoring_scripts)
CREATE TABLE IF NOT EXISTS script_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    script_type VARCHAR(255) NOT NULL,
    script_content TEXT NOT NULL,
    language VARCHAR(50) NOT NULL DEFAULT 'bash',
    default_config JSONB NOT NULL DEFAULT '{}',
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Script policies table: Script configurations at group level (renamed from check_instances)
CREATE TABLE IF NOT EXISTS script_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_from_template_id UUID REFERENCES script_templates(id) ON DELETE SET NULL,
    created_from_template_name VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    script_type VARCHAR(100) NOT NULL,
    script_content TEXT NOT NULL,
    language VARCHAR(50) NOT NULL DEFAULT 'bash',
    scope VARCHAR(20) NOT NULL CHECK (scope IN ('global', 'group')),
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

-- Bootstrap tokens table: Auto-registration tokens (without default_group_id)
CREATE TABLE IF NOT EXISTS bootstrap_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
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

-- Groups indexes
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
CREATE INDEX IF NOT EXISTS idx_alert_rules_created_from_template_id ON alert_rules(created_from_template_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_alert_rules_deleted_at ON alert_rules(deleted_at);

-- Script templates indexes
CREATE INDEX IF NOT EXISTS idx_script_templates_script_type ON script_templates(script_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_script_templates_name ON script_templates(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_script_templates_deleted_at ON script_templates(deleted_at);

-- Script policies indexes
CREATE INDEX IF NOT EXISTS idx_script_policies_created_from_template_id ON script_policies(created_from_template_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_script_policies_scope ON script_policies(scope) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_script_policies_group ON script_policies(group_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_script_policies_deleted_at ON script_policies(deleted_at);

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
