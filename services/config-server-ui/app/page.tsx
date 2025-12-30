'use client'

import { Suspense } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ActiveAlerts } from '@/components/alerts/active-alerts'
import { DashboardStats } from '@/components/dashboard/stats'
import { ExternalLinks } from '@/components/dashboard/external-links'
import { LoadingStats, LoadingCard } from '@/components/shared/loading'

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      {/* Stats Cards */}
      <Suspense fallback={<LoadingStats />}>
        <DashboardStats />
      </Suspense>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Active Alerts */}
        <Card>
          <CardHeader>
            <CardTitle>Active Alerts</CardTitle>
          </CardHeader>
          <CardContent>
            <Suspense fallback={<LoadingCard />}>
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
