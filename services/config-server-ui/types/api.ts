// Common types
export interface TimestampFields {
  created_at: string
  updated_at: string
  deleted_at?: string
}

// Pagination
export interface PaginationParams {
  page?: number
  limit?: number
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

// Target
export interface Target extends TimestampFields {
  id: string
  hostname: string
  ip_address: string
  port: number
  status: 'active' | 'inactive' | 'down'
  labels: Record<string, string>
  groups: GroupSummary[]
}

export interface GroupSummary {
  id: string
  name: string
}

// Group
export interface Group extends TimestampFields {
  id: string
  name: string
  description: string
  priority: number
  is_default_own: boolean
  metadata: Record<string, unknown>
}

// Exporter
export interface Exporter extends TimestampFields {
  id: string
  target_id: string
  type: string
  port: number
  path: string
  enabled: boolean
  target?: Target
}

// Alert Template
export interface AlertTemplate extends TimestampFields {
  id: string
  name: string
  description: string
  severity: 'critical' | 'warning' | 'info'
  query_template: string
  default_config: Record<string, unknown>
}

// Alert Rule
export interface AlertRule extends TimestampFields {
  id: string
  group_id: string
  group?: Group
  name: string
  description: string
  severity: 'critical' | 'warning' | 'info'
  query_template: string
  default_config: Record<string, unknown>
  enabled: boolean
  config: Record<string, unknown>
  merge_strategy: string
  priority: number
  created_from_template_id?: string
  created_from_template_name?: string
}

// Active Alert (from Alertmanager)
export interface ActiveAlert {
  fingerprint: string
  status: string
  labels: Record<string, string>
  annotations: Record<string, string>
  starts_at: string
  generator_url: string
}

export interface ActiveAlertsResponse {
  alerts: ActiveAlert[]
  total: number
}

// Script Template
export interface ScriptTemplate extends TimestampFields {
  id: string
  name: string
  description: string
  script_type: string
  script_content: string
  config_schema: Record<string, unknown>
  hash: string
  enabled: boolean
}

// Script Policy
export interface ScriptPolicy extends TimestampFields {
  id: string
  template_id: string
  template?: ScriptTemplate
  group_id?: string
  group?: Group
  config: Record<string, unknown>
  priority: number
  enabled: boolean
}

// Bootstrap Token
export interface BootstrapToken extends TimestampFields {
  id: string
  name: string
  description: string
  token: string  // Only returned on creation
  group_id: string
  group?: Group
  expires_at: string
  max_uses: number
  use_count: number
}

// Health
export interface HealthResponse {
  status: string
  version: string
  database: string
}

// Prometheus Status
export interface PrometheusStatusResponse {
  reachable: boolean
  healthy: boolean
  status?: Record<string, unknown>
}
