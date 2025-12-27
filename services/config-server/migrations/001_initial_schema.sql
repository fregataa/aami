-- Initial schema for AAMI Config Server
-- Creates all tables for domain models with proper relationships and indexes

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Groups table: Hierarchical organization with namespace
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    namespace VARCHAR(50) NOT NULL CHECK (namespace IN ('infrastructure', 'logical', 'environment')),
    parent_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    description TEXT,
    priority INTEGER NOT NULL DEFAULT 100,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Targets table: Monitored servers/nodes
CREATE TABLE targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hostname VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(45) NOT NULL,
    primary_group_id UUID NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'down')),
    labels JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    last_seen TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Target secondary groups (many-to-many)
CREATE TABLE target_secondary_groups (
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (target_id, group_id)
);

-- Exporters table: Metric collectors configuration
CREATE TABLE exporters (
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
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Alert templates table: Reusable alert definitions
CREATE TABLE alert_templates (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('critical', 'warning', 'info')),
    query_template TEXT NOT NULL,
    default_config JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Alert rules table: Group-specific alert configurations
CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    template_id VARCHAR(100) NOT NULL REFERENCES alert_templates(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB NOT NULL DEFAULT '{}',
    merge_strategy VARCHAR(20) NOT NULL DEFAULT 'override' CHECK (merge_strategy IN ('override', 'merge')),
    priority INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Check settings table: Configuration settings at group level
CREATE TABLE check_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    check_type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    merge_strategy VARCHAR(20) NOT NULL DEFAULT 'merge' CHECK (merge_strategy IN ('override', 'merge')),
    priority INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Bootstrap tokens table: Auto-registration tokens
CREATE TABLE bootstrap_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    default_group_id UUID NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    max_uses INTEGER NOT NULL CHECK (max_uses > 0),
    uses INTEGER NOT NULL DEFAULT 0 CHECK (uses >= 0),
    expires_at TIMESTAMP NOT NULL,
    labels JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_groups_namespace ON groups(namespace);
CREATE INDEX idx_groups_parent ON groups(parent_id);
CREATE INDEX idx_groups_name ON groups(name);

CREATE INDEX idx_targets_primary_group ON targets(primary_group_id);
CREATE INDEX idx_targets_hostname ON targets(hostname);
CREATE INDEX idx_targets_status ON targets(status);

CREATE INDEX idx_target_secondary_groups_target ON target_secondary_groups(target_id);
CREATE INDEX idx_target_secondary_groups_group ON target_secondary_groups(group_id);

CREATE INDEX idx_exporters_target ON exporters(target_id);
CREATE INDEX idx_exporters_type ON exporters(type);

CREATE INDEX idx_alert_rules_group ON alert_rules(group_id);
CREATE INDEX idx_alert_rules_template ON alert_rules(template_id);

CREATE INDEX idx_check_settings_group ON check_settings(group_id);
CREATE INDEX idx_check_settings_type ON check_settings(check_type);

CREATE INDEX idx_bootstrap_tokens_token ON bootstrap_tokens(token);
CREATE INDEX idx_bootstrap_tokens_expires ON bootstrap_tokens(expires_at);

-- Insert default alert templates
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
 '{"threshold": 90, "for": "5m"}'::jsonb);
