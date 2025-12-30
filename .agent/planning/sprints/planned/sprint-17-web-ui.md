# Sprint 17: Web UI Implementation

**Status**: 📋 Planned
**Priority**: High
**Duration**: 4-5 weeks
**Started**: TBD
**Completed**: TBD

## Goals

AAMI Config Server 관리를 위한 웹 기반 사용자 인터페이스를 구현합니다.

### Key Objectives
1. **Dashboard**: 시스템 상태, 노드 요약, 활성 알림 표시
2. **CRUD Operations**: 모든 리소스(Target, Group, Alert Rule 등) 관리
3. **Real-time Updates**: SWR을 활용한 자동 데이터 갱신
4. **Code Editor**: Monaco Editor를 사용한 스크립트 템플릿 편집
5. **Static Export**: nginx로 서빙 가능한 정적 파일 생성

---

## Technology Stack

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| Framework | Next.js | 15.x | React 기반 프레임워크, App Router |
| UI Components | shadcn/ui | latest | Radix UI 기반 컴포넌트 |
| Styling | Tailwind CSS | 4.x | 유틸리티 기반 CSS |
| Data Fetching | SWR | 2.x | 캐싱, 실시간 갱신 |
| Forms | react-hook-form | 7.x | 폼 상태 관리 |
| Validation | zod | 3.x | 스키마 검증 |
| Code Editor | Monaco Editor | @monaco-editor/react | 스크립트 편집 |
| HTTP Client | fetch (native) | - | API 통신 |
| Icons | Lucide React | latest | 아이콘 |

---

## Project Structure

```
services/config-server-ui/
├── app/
│   ├── layout.tsx                 # Root layout (sidebar, header)
│   ├── page.tsx                   # Dashboard
│   ├── targets/
│   │   ├── page.tsx               # Target list
│   │   └── [id]/page.tsx          # Target detail
│   ├── groups/
│   │   ├── page.tsx               # Group list
│   │   └── [id]/page.tsx          # Group detail
│   ├── exporters/
│   │   └── page.tsx               # Exporter list
│   ├── alerts/
│   │   ├── templates/
│   │   │   ├── page.tsx           # Alert template list
│   │   │   └── [id]/page.tsx      # Alert template detail
│   │   └── rules/
│   │       ├── page.tsx           # Alert rule list
│   │       └── [id]/page.tsx      # Alert rule detail
│   ├── scripts/
│   │   ├── templates/
│   │   │   ├── page.tsx           # Script template list
│   │   │   └── [id]/page.tsx      # Script template detail (Monaco)
│   │   └── policies/
│   │       ├── page.tsx           # Script policy list
│   │       └── [id]/page.tsx      # Script policy detail
│   ├── bootstrap/
│   │   └── page.tsx               # Bootstrap token management
│   └── settings/
│       └── page.tsx               # Settings (Prometheus status)
├── components/
│   ├── ui/                        # shadcn/ui components
│   ├── layout/
│   │   ├── sidebar.tsx            # Navigation sidebar
│   │   ├── header.tsx             # Top header with alerts badge
│   │   └── breadcrumb.tsx         # Breadcrumb navigation
│   ├── shared/
│   │   ├── data-table.tsx         # Reusable data table
│   │   ├── delete-dialog.tsx      # Confirm delete dialog
│   │   ├── loading.tsx            # Loading states
│   │   └── error-boundary.tsx     # Error handling
│   ├── targets/
│   │   ├── target-table.tsx
│   │   ├── target-form.tsx
│   │   └── target-card.tsx
│   ├── groups/
│   │   ├── group-table.tsx
│   │   └── group-form.tsx
│   ├── alerts/
│   │   ├── active-alerts.tsx      # Real-time alert list
│   │   ├── alert-badge.tsx        # Alert count badge
│   │   ├── template-form.tsx
│   │   └── rule-form.tsx
│   ├── scripts/
│   │   ├── script-editor.tsx      # Monaco Editor wrapper
│   │   └── policy-form.tsx
│   └── bootstrap/
│       ├── token-table.tsx
│       ├── token-form.tsx
│       └── command-generator.tsx  # Bootstrap command generator
├── lib/
│   ├── api/
│   │   ├── client.ts              # Base API client
│   │   ├── targets.ts             # Target API functions
│   │   ├── groups.ts              # Group API functions
│   │   ├── exporters.ts           # Exporter API functions
│   │   ├── alert-templates.ts     # Alert template API
│   │   ├── alert-rules.ts         # Alert rule API
│   │   ├── script-templates.ts    # Script template API
│   │   ├── script-policies.ts     # Script policy API
│   │   ├── bootstrap-tokens.ts    # Bootstrap token API
│   │   └── prometheus.ts          # Prometheus management API
│   ├── hooks/
│   │   ├── use-targets.ts         # Target SWR hooks
│   │   ├── use-groups.ts          # Group SWR hooks
│   │   ├── use-alerts.ts          # Alert SWR hooks
│   │   └── use-active-alerts.ts   # Real-time alerts hook
│   ├── utils/
│   │   ├── cn.ts                  # className utility
│   │   └── format.ts              # Date/number formatting
│   └── config.ts                  # Environment config
├── types/
│   ├── api.ts                     # API response types
│   ├── target.ts                  # Target types
│   ├── group.ts                   # Group types
│   └── alert.ts                   # Alert types
├── public/
│   └── favicon.ico
├── next.config.ts                 # Next.js config (static export)
├── tailwind.config.ts
├── tsconfig.json
├── package.json
├── Dockerfile
└── nginx.conf
```

---

## Development Phases

### Phase 1: Project Foundation (3-4 days)

#### 1.1 프로젝트 초기화

**Tasks:**
- [ ] Next.js 15 프로젝트 생성 (`create-next-app`)
- [ ] TypeScript strict mode 설정
- [ ] Tailwind CSS 4 설정
- [ ] shadcn/ui 초기화 및 기본 컴포넌트 설치
- [ ] ESLint + Prettier 설정 (Biome 고려)
- [ ] 환경 변수 설정 (`NEXT_PUBLIC_API_URL`)

**Commands:**
```bash
cd services
npx create-next-app@latest config-server-ui --typescript --tailwind --app --src-dir=false
cd config-server-ui
npx shadcn@latest init
npx shadcn@latest add button card dialog dropdown-menu form input label select table tabs toast
```

**`next.config.ts`:**
```typescript
import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  output: 'export',  // Static export for nginx serving
  trailingSlash: true,
  images: {
    unoptimized: true,  // Required for static export
  },
}

export default nextConfig
```

**`lib/config.ts`:**
```typescript
export const config = {
  apiUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
}
```

#### 1.2 API Client 구현

**`lib/api/client.ts`:**
```typescript
import { config } from '@/lib/config'

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

interface ErrorResponse {
  error: {
    code: string
    message: string
    details?: string
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error: ErrorResponse = await response.json().catch(() => ({
      error: { code: 'UNKNOWN', message: 'Unknown error occurred' }
    }))
    throw new ApiError(
      response.status,
      error.error.code,
      error.error.message,
      error.error.details
    )
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return undefined as T
  }

  return response.json()
}

export const api = {
  async get<T>(path: string): Promise<T> {
    const response = await fetch(`${config.apiUrl}${path}`)
    return handleResponse<T>(response)
  },

  async post<T>(path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${config.apiUrl}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: body ? JSON.stringify(body) : undefined,
    })
    return handleResponse<T>(response)
  },

  async put<T>(path: string, body: unknown): Promise<T> {
    const response = await fetch(`${config.apiUrl}${path}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    return handleResponse<T>(response)
  },

  async delete(path: string, body: { id: string }): Promise<void> {
    const response = await fetch(`${config.apiUrl}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    return handleResponse<void>(response)
  },
}
```

#### 1.3 Layout 구현

**`app/layout.tsx`:**
```tsx
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { Sidebar } from '@/components/layout/sidebar'
import { Header } from '@/components/layout/header'
import { Toaster } from '@/components/ui/toaster'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'AAMI Config Server',
  description: 'AAMI Monitoring Configuration Management',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <div className="flex h-screen">
          <Sidebar />
          <div className="flex flex-1 flex-col overflow-hidden">
            <Header />
            <main className="flex-1 overflow-y-auto bg-gray-50 p-6">
              {children}
            </main>
          </div>
        </div>
        <Toaster />
      </body>
    </html>
  )
}
```

**`components/layout/sidebar.tsx`:**
```tsx
'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils/cn'
import {
  LayoutDashboard,
  Server,
  FolderTree,
  Gauge,
  Bell,
  FileCode,
  Key,
  Settings,
} from 'lucide-react'

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Targets', href: '/targets', icon: Server },
  { name: 'Groups', href: '/groups', icon: FolderTree },
  { name: 'Exporters', href: '/exporters', icon: Gauge },
  {
    name: 'Alerts',
    children: [
      { name: 'Templates', href: '/alerts/templates' },
      { name: 'Rules', href: '/alerts/rules' },
    ],
    icon: Bell,
  },
  {
    name: 'Scripts',
    children: [
      { name: 'Templates', href: '/scripts/templates' },
      { name: 'Policies', href: '/scripts/policies' },
    ],
    icon: FileCode,
  },
  { name: 'Bootstrap', href: '/bootstrap', icon: Key },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export function Sidebar() {
  const pathname = usePathname()

  return (
    <div className="flex h-full w-64 flex-col border-r bg-white">
      <div className="flex h-16 items-center border-b px-6">
        <span className="text-xl font-bold">AAMI</span>
      </div>
      <nav className="flex-1 space-y-1 p-4">
        {navigation.map((item) =>
          item.children ? (
            <div key={item.name} className="space-y-1">
              <div className="flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-500">
                <item.icon className="h-4 w-4" />
                {item.name}
              </div>
              {item.children.map((child) => (
                <Link
                  key={child.href}
                  href={child.href}
                  className={cn(
                    'block rounded-md px-3 py-2 pl-9 text-sm',
                    pathname === child.href
                      ? 'bg-gray-100 font-medium text-gray-900'
                      : 'text-gray-600 hover:bg-gray-50'
                  )}
                >
                  {child.name}
                </Link>
              ))}
            </div>
          ) : (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-2 rounded-md px-3 py-2 text-sm',
                pathname === item.href
                  ? 'bg-gray-100 font-medium text-gray-900'
                  : 'text-gray-600 hover:bg-gray-50'
              )}
            >
              <item.icon className="h-4 w-4" />
              {item.name}
            </Link>
          )
        )}
      </nav>
    </div>
  )
}
```

**`components/layout/header.tsx`:**
```tsx
'use client'

import { AlertBadge } from '@/components/alerts/alert-badge'
import { Breadcrumb } from '@/components/layout/breadcrumb'

export function Header() {
  return (
    <header className="flex h-16 items-center justify-between border-b bg-white px-6">
      <Breadcrumb />
      <div className="flex items-center gap-4">
        <AlertBadge />
      </div>
    </header>
  )
}
```

#### 1.4 Type Definitions

**`types/api.ts`:**
```typescript
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
export interface PrometheusStatus {
  status: string
  url: string
  // Add more fields as needed
}
```

---

### Phase 2: Dashboard & Core Pages (4-5 days)

#### 2.1 Dashboard 구현

**`app/page.tsx`:**
```tsx
import { Suspense } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ActiveAlerts } from '@/components/alerts/active-alerts'
import { DashboardStats } from '@/components/dashboard/stats'
import { ExternalLinks } from '@/components/dashboard/external-links'
import { Skeleton } from '@/components/ui/skeleton'

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      {/* Stats Cards */}
      <Suspense fallback={<StatsLoading />}>
        <DashboardStats />
      </Suspense>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Active Alerts */}
        <Card>
          <CardHeader>
            <CardTitle>Active Alerts</CardTitle>
          </CardHeader>
          <CardContent>
            <Suspense fallback={<Skeleton className="h-48" />}>
              <ActiveAlerts />
            </Suspense>
          </CardContent>
        </Card>

        {/* External Links */}
        <Card>
          <CardHeader>
            <CardTitle>External Services</CardTitle>
          </CardHeader>
          <CardContent>
            <ExternalLinks />
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function StatsLoading() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {[1, 2, 3, 4].map((i) => (
        <Skeleton key={i} className="h-24" />
      ))}
    </div>
  )
}
```

**`components/dashboard/stats.tsx`:**
```tsx
'use client'

import useSWR from 'swr'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { api } from '@/lib/api/client'
import { Server, FolderTree, Bell, Activity } from 'lucide-react'
import type { Target, Group, ActiveAlertsResponse, HealthResponse } from '@/types/api'

export function DashboardStats() {
  const { data: targets } = useSWR<Target[]>('/api/v1/targets', api.get)
  const { data: groups } = useSWR<Group[]>('/api/v1/groups', api.get)
  const { data: alerts } = useSWR<ActiveAlertsResponse>(
    '/api/v1/alerts/active',
    api.get,
    { refreshInterval: 10000 }  // Refresh every 10 seconds
  )
  const { data: health } = useSWR<HealthResponse>('/health', api.get)

  const stats = [
    {
      title: 'Total Targets',
      value: targets?.length ?? '-',
      icon: Server,
      description: `${targets?.filter(t => t.status === 'active').length ?? 0} active`,
    },
    {
      title: 'Groups',
      value: groups?.length ?? '-',
      icon: FolderTree,
    },
    {
      title: 'Active Alerts',
      value: alerts?.total ?? '-',
      icon: Bell,
      highlight: (alerts?.total ?? 0) > 0,
    },
    {
      title: 'System Status',
      value: health?.status === 'healthy' ? 'Healthy' : 'Unknown',
      icon: Activity,
      description: health?.version,
    },
  ]

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {stats.map((stat) => (
        <Card key={stat.title}>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              {stat.title}
            </CardTitle>
            <stat.icon className="h-4 w-4 text-gray-400" />
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${stat.highlight ? 'text-red-600' : ''}`}>
              {stat.value}
            </div>
            {stat.description && (
              <p className="text-xs text-gray-500">{stat.description}</p>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
```

**`components/alerts/active-alerts.tsx`:**
```tsx
'use client'

import useSWR from 'swr'
import { api } from '@/lib/api/client'
import { Badge } from '@/components/ui/badge'
import { formatDistanceToNow } from 'date-fns'
import type { ActiveAlertsResponse } from '@/types/api'

export function ActiveAlerts() {
  const { data, isLoading } = useSWR<ActiveAlertsResponse>(
    '/api/v1/alerts/active',
    api.get,
    { refreshInterval: 10000 }
  )

  if (isLoading) {
    return <div className="text-gray-500">Loading alerts...</div>
  }

  if (!data?.alerts.length) {
    return (
      <div className="flex h-32 items-center justify-center text-gray-500">
        No active alerts
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {data.alerts.slice(0, 5).map((alert) => (
        <div
          key={alert.fingerprint}
          className="flex items-center justify-between rounded-lg border p-3"
        >
          <div className="flex items-center gap-3">
            <SeverityBadge severity={alert.labels.severity} />
            <div>
              <div className="font-medium">{alert.labels.alertname}</div>
              <div className="text-sm text-gray-500">
                {alert.labels.instance}
              </div>
            </div>
          </div>
          <div className="text-sm text-gray-500">
            {formatDistanceToNow(new Date(alert.starts_at), { addSuffix: true })}
          </div>
        </div>
      ))}
      {data.total > 5 && (
        <div className="text-center text-sm text-gray-500">
          and {data.total - 5} more alerts
        </div>
      )}
    </div>
  )
}

function SeverityBadge({ severity }: { severity?: string }) {
  const variant = severity === 'critical' ? 'destructive'
    : severity === 'warning' ? 'warning'
    : 'secondary'

  return (
    <Badge variant={variant}>
      {severity?.toUpperCase() || 'UNKNOWN'}
    </Badge>
  )
}
```

#### 2.2 Target 관리 페이지

**`lib/api/targets.ts`:**
```typescript
import { api } from './client'
import type { Target } from '@/types/api'

export const targetsApi = {
  list: () => api.get<Target[]>('/api/v1/targets'),

  getById: (id: string) => api.get<Target>(`/api/v1/targets/${id}`),

  getByHostname: (hostname: string) =>
    api.get<Target>(`/api/v1/targets/hostname/${hostname}`),

  getByGroup: (groupId: string) =>
    api.get<Target[]>(`/api/v1/targets/group/${groupId}`),

  create: (data: CreateTargetRequest) =>
    api.post<Target>('/api/v1/targets', data),

  update: (id: string, data: UpdateTargetRequest) =>
    api.put<Target>(`/api/v1/targets/${id}`, data),

  delete: (id: string) =>
    api.delete('/api/v1/targets/delete', { id }),

  restore: (id: string) =>
    api.post('/api/v1/targets/restore', { id }),

  purge: (id: string) =>
    api.post('/api/v1/targets/purge', { id }),
}

export interface CreateTargetRequest {
  hostname: string
  ip_address: string
  port?: number
  labels?: Record<string, string>
  group_ids?: string[]
}

export interface UpdateTargetRequest {
  hostname?: string
  ip_address?: string
  port?: number
  labels?: Record<string, string>
  group_ids?: string[]
}
```

**`lib/hooks/use-targets.ts`:**
```typescript
import useSWR from 'swr'
import { targetsApi } from '@/lib/api/targets'

export function useTargets() {
  const { data, error, isLoading, mutate } = useSWR(
    '/api/v1/targets',
    targetsApi.list
  )

  return {
    targets: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useTarget(id: string) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/targets/${id}` : null,
    () => targetsApi.getById(id)
  )

  return {
    target: data,
    isLoading,
    error,
    mutate,
  }
}
```

**`app/targets/page.tsx`:**
```tsx
'use client'

import { useState } from 'react'
import { useTargets } from '@/lib/hooks/use-targets'
import { targetsApi } from '@/lib/api/targets'
import { Button } from '@/components/ui/button'
import { TargetTable } from '@/components/targets/target-table'
import { TargetForm } from '@/components/targets/target-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { Plus } from 'lucide-react'
import { useToast } from '@/components/ui/use-toast'
import type { Target } from '@/types/api'

export default function TargetsPage() {
  const { targets, isLoading, mutate } = useTargets()
  const [showForm, setShowForm] = useState(false)
  const [editingTarget, setEditingTarget] = useState<Target | null>(null)
  const [deletingTarget, setDeletingTarget] = useState<Target | null>(null)
  const { toast } = useToast()

  const handleCreate = async (data: CreateTargetRequest) => {
    try {
      await targetsApi.create(data)
      toast({ title: 'Target created successfully' })
      mutate()
      setShowForm(false)
    } catch (error) {
      toast({ title: 'Failed to create target', variant: 'destructive' })
    }
  }

  const handleUpdate = async (data: UpdateTargetRequest) => {
    if (!editingTarget) return
    try {
      await targetsApi.update(editingTarget.id, data)
      toast({ title: 'Target updated successfully' })
      mutate()
      setEditingTarget(null)
    } catch (error) {
      toast({ title: 'Failed to update target', variant: 'destructive' })
    }
  }

  const handleDelete = async () => {
    if (!deletingTarget) return
    try {
      await targetsApi.delete(deletingTarget.id)
      toast({ title: 'Target deleted successfully' })
      mutate()
      setDeletingTarget(null)
    } catch (error) {
      toast({ title: 'Failed to delete target', variant: 'destructive' })
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Targets</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Target
        </Button>
      </div>

      <TargetTable
        targets={targets}
        isLoading={isLoading}
        onEdit={setEditingTarget}
        onDelete={setDeletingTarget}
      />

      <TargetForm
        open={showForm || !!editingTarget}
        onClose={() => {
          setShowForm(false)
          setEditingTarget(null)
        }}
        target={editingTarget}
        onSubmit={editingTarget ? handleUpdate : handleCreate}
      />

      <DeleteDialog
        open={!!deletingTarget}
        onClose={() => setDeletingTarget(null)}
        onConfirm={handleDelete}
        title="Delete Target"
        description={`Are you sure you want to delete "${deletingTarget?.hostname}"?`}
      />
    </div>
  )
}
```

**`components/targets/target-table.tsx`:**
```tsx
'use client'

import Link from 'next/link'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { MoreHorizontal, Edit, Trash2 } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import type { Target } from '@/types/api'

interface TargetTableProps {
  targets: Target[]
  isLoading: boolean
  onEdit: (target: Target) => void
  onDelete: (target: Target) => void
}

export function TargetTable({ targets, isLoading, onEdit, onDelete }: TargetTableProps) {
  if (isLoading) {
    return <TableLoading />
  }

  if (!targets.length) {
    return (
      <div className="flex h-32 items-center justify-center rounded-lg border text-gray-500">
        No targets found. Create your first target to get started.
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Hostname</TableHead>
            <TableHead>IP Address</TableHead>
            <TableHead>Port</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Groups</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {targets.map((target) => (
            <TableRow key={target.id}>
              <TableCell>
                <Link
                  href={`/targets/${target.id}`}
                  className="font-medium hover:underline"
                >
                  {target.hostname}
                </Link>
              </TableCell>
              <TableCell>{target.ip_address}</TableCell>
              <TableCell>{target.port}</TableCell>
              <TableCell>
                <StatusBadge status={target.status} />
              </TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  {target.groups?.map((group) => (
                    <Badge key={group.id} variant="outline">
                      {group.name}
                    </Badge>
                  ))}
                </div>
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => onEdit(target)}>
                      <Edit className="mr-2 h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => onDelete(target)}
                      className="text-red-600"
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

function StatusBadge({ status }: { status: Target['status'] }) {
  const variant = status === 'active' ? 'success'
    : status === 'inactive' ? 'secondary'
    : 'destructive'

  return <Badge variant={variant}>{status}</Badge>
}

function TableLoading() {
  return (
    <div className="space-y-3">
      {[1, 2, 3, 4, 5].map((i) => (
        <Skeleton key={i} className="h-12 w-full" />
      ))}
    </div>
  )
}
```

**`components/targets/target-form.tsx`:**
```tsx
'use client'

import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { GroupMultiSelect } from '@/components/groups/group-multi-select'
import type { Target } from '@/types/api'

const formSchema = z.object({
  hostname: z.string().min(1, 'Hostname is required'),
  ip_address: z.string().ip('Invalid IP address'),
  port: z.number().min(1).max(65535).optional(),
  group_ids: z.array(z.string()).optional(),
})

type FormValues = z.infer<typeof formSchema>

interface TargetFormProps {
  open: boolean
  onClose: () => void
  target?: Target | null
  onSubmit: (data: FormValues) => Promise<void>
}

export function TargetForm({ open, onClose, target, onSubmit }: TargetFormProps) {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      hostname: '',
      ip_address: '',
      port: 9100,
      group_ids: [],
    },
  })

  useEffect(() => {
    if (target) {
      form.reset({
        hostname: target.hostname,
        ip_address: target.ip_address,
        port: target.port,
        group_ids: target.groups?.map((g) => g.id) ?? [],
      })
    } else {
      form.reset({
        hostname: '',
        ip_address: '',
        port: 9100,
        group_ids: [],
      })
    }
  }, [target, form])

  const handleSubmit = async (data: FormValues) => {
    await onSubmit(data)
    form.reset()
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{target ? 'Edit Target' : 'Create Target'}</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="hostname"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Hostname</FormLabel>
                  <FormControl>
                    <Input placeholder="server-01.example.com" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="ip_address"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>IP Address</FormLabel>
                  <FormControl>
                    <Input placeholder="192.168.1.100" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="port"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Port</FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      placeholder="9100"
                      {...field}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="group_ids"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Groups</FormLabel>
                  <FormControl>
                    <GroupMultiSelect
                      value={field.value ?? []}
                      onChange={field.onChange}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={form.formState.isSubmitting}>
                {form.formState.isSubmitting ? 'Saving...' : 'Save'}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
```

---

### Phase 3: Alert & Script Management (4-5 days)

#### 3.1 Alert Template 관리

**`app/alerts/templates/[id]/page.tsx`:**
```tsx
'use client'

import { use } from 'react'
import { useRouter } from 'next/navigation'
import useSWR from 'swr'
import { api } from '@/lib/api/client'
import { alertTemplatesApi } from '@/lib/api/alert-templates'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertTemplateForm } from '@/components/alerts/template-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { useToast } from '@/components/ui/use-toast'
import { Edit, Trash2, ArrowLeft } from 'lucide-react'
import Link from 'next/link'
import type { AlertTemplate, AlertRule } from '@/types/api'

export default function AlertTemplateDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  const router = useRouter()
  const { toast } = useToast()

  const { data: template, isLoading, mutate } = useSWR<AlertTemplate>(
    `/api/v1/alert-templates/${id}`,
    () => alertTemplatesApi.getById(id)
  )

  const { data: rules } = useSWR<AlertRule[]>(
    `/api/v1/alert-rules/template/${id}`,
    () => api.get(`/api/v1/alert-rules/template/${id}`)
  )

  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)

  const handleUpdate = async (data: UpdateAlertTemplateRequest) => {
    try {
      await alertTemplatesApi.update(id, data)
      toast({ title: 'Template updated successfully' })
      mutate()
      setShowEdit(false)
    } catch (error) {
      toast({ title: 'Failed to update template', variant: 'destructive' })
    }
  }

  const handleDelete = async () => {
    try {
      await alertTemplatesApi.delete(id)
      toast({ title: 'Template deleted successfully' })
      router.push('/alerts/templates')
    } catch (error) {
      toast({ title: 'Failed to delete template', variant: 'destructive' })
    }
  }

  if (isLoading) {
    return <Skeleton className="h-96" />
  }

  if (!template) {
    return <div>Template not found</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/alerts/templates">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <h1 className="text-2xl font-bold">{template.name}</h1>
        <SeverityBadge severity={template.severity} />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Template Details</CardTitle>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => setShowEdit(true)}>
                <Edit className="mr-2 h-4 w-4" />
                Edit
              </Button>
              <Button variant="outline" size="sm" onClick={() => setShowDelete(true)}>
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="text-sm font-medium text-gray-500">Description</div>
              <div>{template.description || '-'}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Query Template</div>
              <pre className="mt-1 rounded bg-gray-100 p-3 text-sm">
                {template.query_template}
              </pre>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Default Config</div>
              <pre className="mt-1 rounded bg-gray-100 p-3 text-sm">
                {JSON.stringify(template.default_config, null, 2)}
              </pre>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Rules Using This Template</CardTitle>
          </CardHeader>
          <CardContent>
            {rules?.length ? (
              <div className="space-y-2">
                {rules.map((rule) => (
                  <Link
                    key={rule.id}
                    href={`/alerts/rules/${rule.id}`}
                    className="block rounded border p-3 hover:bg-gray-50"
                  >
                    <div className="font-medium">{rule.name}</div>
                    <div className="text-sm text-gray-500">
                      Group: {rule.group?.name ?? 'Unknown'}
                    </div>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="text-gray-500">No rules using this template</div>
            )}
          </CardContent>
        </Card>
      </div>

      <AlertTemplateForm
        open={showEdit}
        onClose={() => setShowEdit(false)}
        template={template}
        onSubmit={handleUpdate}
      />

      <DeleteDialog
        open={showDelete}
        onClose={() => setShowDelete(false)}
        onConfirm={handleDelete}
        title="Delete Template"
        description={`Are you sure you want to delete "${template.name}"?`}
      />
    </div>
  )
}
```

#### 3.2 Script Template with Monaco Editor

**`components/scripts/script-editor.tsx`:**
```tsx
'use client'

import { useRef, useCallback } from 'react'
import Editor, { OnMount } from '@monaco-editor/react'
import type { editor } from 'monaco-editor'

interface ScriptEditorProps {
  value: string
  onChange: (value: string) => void
  language?: string
  readOnly?: boolean
  height?: string
}

export function ScriptEditor({
  value,
  onChange,
  language = 'shell',
  readOnly = false,
  height = '400px',
}: ScriptEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null)

  const handleMount: OnMount = useCallback((editor) => {
    editorRef.current = editor
  }, [])

  const handleChange = useCallback(
    (value: string | undefined) => {
      onChange(value ?? '')
    },
    [onChange]
  )

  return (
    <div className="overflow-hidden rounded-lg border">
      <Editor
        height={height}
        language={language}
        value={value}
        onChange={handleChange}
        onMount={handleMount}
        options={{
          readOnly,
          minimap: { enabled: false },
          lineNumbers: 'on',
          scrollBeyondLastLine: false,
          fontSize: 14,
          tabSize: 2,
          automaticLayout: true,
          wordWrap: 'on',
        }}
        theme="vs-light"
      />
    </div>
  )
}
```

**`app/scripts/templates/[id]/page.tsx`:**
```tsx
'use client'

import { use, useState } from 'react'
import { useRouter } from 'next/navigation'
import useSWR from 'swr'
import { scriptTemplatesApi } from '@/lib/api/script-templates'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { ScriptEditor } from '@/components/scripts/script-editor'
import { useToast } from '@/components/ui/use-toast'
import { Save, RotateCcw, CheckCircle, XCircle } from 'lucide-react'
import type { ScriptTemplate } from '@/types/api'

export default function ScriptTemplateDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  const { toast } = useToast()

  const { data: template, isLoading, mutate } = useSWR<ScriptTemplate>(
    `/api/v1/script-templates/${id}`,
    () => scriptTemplatesApi.getById(id)
  )

  const [editedContent, setEditedContent] = useState<string | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const [hashValid, setHashValid] = useState<boolean | null>(null)

  const hasChanges = editedContent !== null && editedContent !== template?.script_content

  const handleSave = async () => {
    if (!hasChanges || editedContent === null) return

    setIsSaving(true)
    try {
      await scriptTemplatesApi.update(id, { script_content: editedContent })
      toast({ title: 'Script saved successfully' })
      mutate()
      setEditedContent(null)
    } catch (error) {
      toast({ title: 'Failed to save script', variant: 'destructive' })
    } finally {
      setIsSaving(false)
    }
  }

  const handleReset = () => {
    setEditedContent(null)
  }

  const handleVerifyHash = async () => {
    try {
      const result = await scriptTemplatesApi.verifyHash(id)
      setHashValid(result.valid)
      toast({
        title: result.valid ? 'Hash is valid' : 'Hash mismatch detected',
        variant: result.valid ? 'default' : 'destructive',
      })
    } catch (error) {
      toast({ title: 'Failed to verify hash', variant: 'destructive' })
    }
  }

  if (isLoading) {
    return <Skeleton className="h-96" />
  }

  if (!template) {
    return <div>Template not found</div>
  }

  const currentContent = editedContent ?? template.script_content

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h1 className="text-2xl font-bold">{template.name}</h1>
          <Badge>{template.script_type}</Badge>
          {template.enabled ? (
            <Badge variant="success">Enabled</Badge>
          ) : (
            <Badge variant="secondary">Disabled</Badge>
          )}
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleVerifyHash}>
            {hashValid === true && <CheckCircle className="mr-2 h-4 w-4 text-green-600" />}
            {hashValid === false && <XCircle className="mr-2 h-4 w-4 text-red-600" />}
            Verify Hash
          </Button>
          {hasChanges && (
            <>
              <Button variant="outline" onClick={handleReset}>
                <RotateCcw className="mr-2 h-4 w-4" />
                Reset
              </Button>
              <Button onClick={handleSave} disabled={isSaving}>
                <Save className="mr-2 h-4 w-4" />
                {isSaving ? 'Saving...' : 'Save'}
              </Button>
            </>
          )}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Script Content</CardTitle>
        </CardHeader>
        <CardContent>
          <ScriptEditor
            value={currentContent}
            onChange={setEditedContent}
            language="shell"
            height="500px"
          />
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Details</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="text-sm font-medium text-gray-500">Description</div>
              <div>{template.description || '-'}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Hash</div>
              <code className="text-sm">{template.hash}</code>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Config Schema</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="rounded bg-gray-100 p-3 text-sm">
              {JSON.stringify(template.config_schema, null, 2)}
            </pre>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
```

---

### Phase 4: Bootstrap Token & Settings (3-4 days)

#### 4.1 Bootstrap Command Generator

**`components/bootstrap/command-generator.tsx`:**
```tsx
'use client'

import { useState, useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Copy, Check } from 'lucide-react'
import { useToast } from '@/components/ui/use-toast'
import type { BootstrapToken } from '@/types/api'

interface CommandGeneratorProps {
  tokens: BootstrapToken[]
}

export function CommandGenerator({ tokens }: CommandGeneratorProps) {
  const [selectedTokenId, setSelectedTokenId] = useState<string>('')
  const [serverUrl, setServerUrl] = useState('https://config.example.com')
  const [labels, setLabels] = useState('')
  const [copied, setCopied] = useState(false)
  const { toast } = useToast()

  const selectedToken = tokens.find((t) => t.id === selectedTokenId)

  const command = useMemo(() => {
    if (!selectedToken) return ''

    let cmd = `curl -fsSL ${serverUrl}/api/v1/bootstrap/script | sudo bash -s -- \\
  --token ${selectedToken.token || '<token>'} \\
  --server ${serverUrl}`

    if (labels.trim()) {
      const labelParts = labels.split(',').map((l) => l.trim()).filter(Boolean)
      labelParts.forEach((label) => {
        cmd += ` \\
  --labels ${label}`
      })
    }

    return cmd
  }, [selectedToken, serverUrl, labels])

  const handleCopy = async () => {
    if (!command) return

    try {
      await navigator.clipboard.writeText(command)
      setCopied(true)
      toast({ title: 'Command copied to clipboard' })
      setTimeout(() => setCopied(false), 2000)
    } catch (error) {
      toast({ title: 'Failed to copy', variant: 'destructive' })
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Bootstrap Command Generator</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label>Config Server URL</Label>
            <Input
              value={serverUrl}
              onChange={(e) => setServerUrl(e.target.value)}
              placeholder="https://config.example.com"
            />
          </div>
          <div className="space-y-2">
            <Label>Bootstrap Token</Label>
            <Select value={selectedTokenId} onValueChange={setSelectedTokenId}>
              <SelectTrigger>
                <SelectValue placeholder="Select a token" />
              </SelectTrigger>
              <SelectContent>
                {tokens.map((token) => (
                  <SelectItem key={token.id} value={token.id}>
                    {token.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="space-y-2">
          <Label>Labels (comma-separated, e.g., env=prod,rack=A1)</Label>
          <Input
            value={labels}
            onChange={(e) => setLabels(e.target.value)}
            placeholder="env=production, datacenter=dc1"
          />
        </div>

        {command && (
          <div className="space-y-2">
            <Label>Generated Command</Label>
            <div className="relative">
              <pre className="overflow-x-auto rounded-lg bg-gray-900 p-4 text-sm text-gray-100">
                {command}
              </pre>
              <Button
                variant="ghost"
                size="icon"
                className="absolute right-2 top-2 text-gray-400 hover:text-white"
                onClick={handleCopy}
              >
                {copied ? (
                  <Check className="h-4 w-4" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
```

#### 4.2 Settings Page

**`app/settings/page.tsx`:**
```tsx
'use client'

import useSWR from 'swr'
import { api } from '@/lib/api/client'
import { prometheusApi } from '@/lib/api/prometheus'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { useToast } from '@/components/ui/use-toast'
import { RefreshCw, ExternalLink, RotateCcw } from 'lucide-react'
import type { PrometheusStatus, HealthResponse } from '@/types/api'

export default function SettingsPage() {
  const { toast } = useToast()

  const { data: health } = useSWR<HealthResponse>('/health', api.get)
  const { data: prometheusStatus, mutate: mutatePrometheus } = useSWR<PrometheusStatus>(
    '/api/v1/prometheus/status',
    () => prometheusApi.getStatus()
  )

  const [isRegenerating, setIsRegenerating] = useState(false)
  const [isReloading, setIsReloading] = useState(false)

  const handleRegenerateRules = async () => {
    setIsRegenerating(true)
    try {
      await prometheusApi.regenerateAllRules()
      toast({ title: 'Rules regenerated successfully' })
    } catch (error) {
      toast({ title: 'Failed to regenerate rules', variant: 'destructive' })
    } finally {
      setIsRegenerating(false)
    }
  }

  const handleReloadPrometheus = async () => {
    setIsReloading(true)
    try {
      await prometheusApi.reload()
      toast({ title: 'Prometheus reloaded successfully' })
      mutatePrometheus()
    } catch (error) {
      toast({ title: 'Failed to reload Prometheus', variant: 'destructive' })
    } finally {
      setIsReloading(false)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Settings</h1>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* System Status */}
        <Card>
          <CardHeader>
            <CardTitle>System Status</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span>Config Server</span>
              <Badge variant={health?.status === 'healthy' ? 'success' : 'destructive'}>
                {health?.status ?? 'Unknown'}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <span>Version</span>
              <span className="text-gray-500">{health?.version ?? '-'}</span>
            </div>
            <div className="flex items-center justify-between">
              <span>Database</span>
              <Badge variant={health?.database === 'connected' ? 'success' : 'destructive'}>
                {health?.database ?? 'Unknown'}
              </Badge>
            </div>
          </CardContent>
        </Card>

        {/* Prometheus Management */}
        <Card>
          <CardHeader>
            <CardTitle>Prometheus Management</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span>Status</span>
              <Badge variant={prometheusStatus?.status === 'connected' ? 'success' : 'secondary'}>
                {prometheusStatus?.status ?? 'Unknown'}
              </Badge>
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                onClick={handleRegenerateRules}
                disabled={isRegenerating}
              >
                <RotateCcw className={`mr-2 h-4 w-4 ${isRegenerating ? 'animate-spin' : ''}`} />
                Regenerate Rules
              </Button>
              <Button
                variant="outline"
                onClick={handleReloadPrometheus}
                disabled={isReloading}
              >
                <RefreshCw className={`mr-2 h-4 w-4 ${isReloading ? 'animate-spin' : ''}`} />
                Reload Prometheus
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* External Links */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>External Services</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-3">
              <a
                href="http://localhost:3000"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center justify-between rounded-lg border p-4 hover:bg-gray-50"
              >
                <span className="font-medium">Grafana</span>
                <ExternalLink className="h-4 w-4 text-gray-400" />
              </a>
              <a
                href="http://localhost:9090"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center justify-between rounded-lg border p-4 hover:bg-gray-50"
              >
                <span className="font-medium">Prometheus</span>
                <ExternalLink className="h-4 w-4 text-gray-400" />
              </a>
              <a
                href="http://localhost:9093"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center justify-between rounded-lg border p-4 hover:bg-gray-50"
              >
                <span className="font-medium">Alertmanager</span>
                <ExternalLink className="h-4 w-4 text-gray-400" />
              </a>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
```

---

### Phase 5: Polish & Deployment (3-4 days)

#### 5.1 Error Handling

**`components/shared/error-boundary.tsx`:**
```tsx
'use client'

import { Component, ReactNode } from 'react'
import { Button } from '@/components/ui/button'
import { AlertTriangle } from 'lucide-react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div className="flex flex-col items-center justify-center gap-4 p-8">
          <AlertTriangle className="h-12 w-12 text-red-500" />
          <h2 className="text-lg font-semibold">Something went wrong</h2>
          <p className="text-gray-500">{this.state.error?.message}</p>
          <Button onClick={() => this.setState({ hasError: false })}>
            Try again
          </Button>
        </div>
      )
    }

    return this.props.children
  }
}
```

#### 5.2 Docker Build

**`Dockerfile`:**
```dockerfile
# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install dependencies
RUN corepack enable && pnpm install --frozen-lockfile

# Copy source files
COPY . .

# Build with static export
ENV NEXT_TELEMETRY_DISABLED=1
RUN pnpm build

# Production stage
FROM nginx:alpine

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy static files from builder
COPY --from=builder /app/out /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

**`nginx.conf`:**
```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # Enable gzip
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;

    # Handle client-side routing
    location / {
        try_files $uri $uri/ $uri.html /index.html;
    }

    # Cache static assets
    location /_next/static/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # API proxy (if needed)
    location /api/ {
        proxy_pass http://config-server:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

---

## Implementation Checklist

### Phase 1: Foundation
- [ ] Next.js 프로젝트 생성
- [ ] shadcn/ui 설정
- [ ] Tailwind CSS 4 설정
- [ ] API Client 구현
- [ ] Type definitions
- [ ] Layout (Sidebar, Header)
- [ ] 기본 컴포넌트 (DataTable, DeleteDialog, Loading)

### Phase 2: Core Pages
- [ ] Dashboard (Stats, Active Alerts, External Links)
- [ ] Target 목록/상세/CRUD
- [ ] Group 목록/상세/CRUD
- [ ] Exporter 목록/CRUD

### Phase 3: Alert & Script
- [ ] Alert Template 목록/상세/CRUD
- [ ] Alert Rule 목록/상세/CRUD
- [ ] Script Template 목록/상세/CRUD (Monaco Editor)
- [ ] Script Policy 목록/상세/CRUD

### Phase 4: Bootstrap & Settings
- [ ] Bootstrap Token 목록/생성/삭제
- [ ] Bootstrap Command Generator
- [ ] Settings (Prometheus 상태/관리)

### Phase 5: Polish
- [ ] Error Boundary
- [ ] Loading States (Skeleton)
- [ ] Toast Notifications
- [ ] Responsive Design
- [ ] Docker Build
- [ ] nginx 설정

---

## Success Criteria

### Functionality
- [ ] 모든 리소스 CRUD 작동
- [ ] Real-time 알림 표시 (10초 갱신)
- [ ] Monaco Editor로 스크립트 편집 가능
- [ ] Bootstrap 명령 생성 및 복사

### Performance
- [ ] 초기 로드 3초 이내
- [ ] 페이지 전환 500ms 이내
- [ ] 정적 파일 gzip 압축

### Compatibility
- [ ] Chrome, Firefox, Safari, Edge 최신 버전 지원
- [ ] 1280px 이상 화면 최적화
- [ ] 768px 이상 반응형

### Deployment
- [ ] Docker 이미지 빌드 성공
- [ ] nginx 서빙 가능
- [ ] Config Server CORS 호환

---

## Dependencies

### External
- Config Server API (기존 완료)
- `GET /api/v1/alerts/active` (완료)

### Optional Enhancements (Post-Sprint)
- Target 검색/필터 파라미터
- 일관된 Pagination 응답 형식

---

## Notes

- Static Export 사용으로 SSR 불필요
- API URL은 환경변수로 주입
- Authentication은 이 스프린트에서 제외
- i18n은 영어만 지원 (초기 버전)

---

## Related Documents

- [Web UI Spec](../../docs/web-ui-spec.md) - 상세 UI 명세
- [API Reference](../../docs/en/API.md) - API 문서
