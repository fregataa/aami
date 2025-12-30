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
    { refreshInterval: 10000 }
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
