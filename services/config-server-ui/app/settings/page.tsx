'use client'

import { useState } from 'react'
import useSWR from 'swr'
import { api } from '@/lib/api/client'
import { prometheusApi } from '@/lib/api/prometheus'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'
import {
  RefreshCw,
  ExternalLink,
  RotateCcw,
  Activity,
  Database,
  Server,
  CheckCircle2,
  XCircle,
} from 'lucide-react'
import type { HealthResponse, PrometheusStatusResponse } from '@/types/api'

export default function SettingsPage() {
  const { data: health, isLoading: healthLoading } = useSWR<HealthResponse>(
    '/health',
    () => api.get('/health'),
    { refreshInterval: 30000 }
  )

  const { data: prometheusStatus, mutate: mutatePrometheus, isLoading: prometheusLoading } = useSWR<PrometheusStatusResponse>(
    '/api/v1/prometheus/status',
    () => prometheusApi.getStatus(),
    { refreshInterval: 30000 }
  )

  const [isRegenerating, setIsRegenerating] = useState(false)
  const [isReloading, setIsReloading] = useState(false)

  const handleRegenerateRules = async () => {
    setIsRegenerating(true)
    try {
      const result = await prometheusApi.regenerateAllRules()
      toast.success(result.message || 'Rules regenerated successfully')
    } catch {
      toast.error('Failed to regenerate rules')
    } finally {
      setIsRegenerating(false)
    }
  }

  const handleReloadPrometheus = async () => {
    setIsReloading(true)
    try {
      const result = await prometheusApi.reload()
      if (result.success) {
        toast.success('Prometheus reloaded successfully')
      } else {
        toast.error(result.message || 'Failed to reload Prometheus')
      }
      mutatePrometheus()
    } catch {
      toast.error('Failed to reload Prometheus')
    } finally {
      setIsReloading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-gray-500">System status and configuration</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* System Status */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5" />
              Config Server Status
            </CardTitle>
            <CardDescription>
              Current status of the AAMI Config Server
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {healthLoading ? (
              <div className="space-y-3">
                <Skeleton className="h-6 w-full" />
                <Skeleton className="h-6 w-full" />
                <Skeleton className="h-6 w-full" />
              </div>
            ) : (
              <>
                <StatusRow
                  label="Server Status"
                  status={health?.status === 'healthy'}
                  value={health?.status || 'Unknown'}
                />
                <StatusRow
                  label="Database"
                  status={health?.database === 'connected'}
                  value={health?.database || 'Unknown'}
                />
                <div className="flex items-center justify-between py-2">
                  <span className="text-sm text-gray-600">Version</span>
                  <span className="font-mono text-sm">{health?.version || 'Unknown'}</span>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Prometheus Status */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Prometheus Status
            </CardTitle>
            <CardDescription>
              Prometheus connection and health status
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {prometheusLoading ? (
              <div className="space-y-3">
                <Skeleton className="h-6 w-full" />
                <Skeleton className="h-6 w-full" />
              </div>
            ) : (
              <>
                <StatusRow
                  label="Reachable"
                  status={prometheusStatus?.reachable}
                  value={prometheusStatus?.reachable ? 'Yes' : 'No'}
                />
                <StatusRow
                  label="Healthy"
                  status={prometheusStatus?.healthy}
                  value={prometheusStatus?.healthy ? 'Yes' : 'No'}
                />
              </>
            )}

            <div className="flex gap-2 pt-2">
              <Button
                variant="outline"
                size="sm"
                onClick={handleRegenerateRules}
                disabled={isRegenerating}
              >
                <RotateCcw className={`mr-2 h-4 w-4 ${isRegenerating ? 'animate-spin' : ''}`} />
                Regenerate Rules
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={handleReloadPrometheus}
                disabled={isReloading}
              >
                <RefreshCw className={`mr-2 h-4 w-4 ${isReloading ? 'animate-spin' : ''}`} />
                Reload
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* External Services */}
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Database className="h-5 w-5" />
              External Services
            </CardTitle>
            <CardDescription>
              Links to related monitoring services
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-3">
              <ExternalServiceLink
                name="Grafana"
                description="Dashboards and visualization"
                url="http://localhost:3000"
              />
              <ExternalServiceLink
                name="Prometheus"
                description="Metrics and queries"
                url="http://localhost:9090"
              />
              <ExternalServiceLink
                name="Alertmanager"
                description="Alert routing and silencing"
                url="http://localhost:9093"
              />
            </div>
          </CardContent>
        </Card>

        {/* API Information */}
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle>API Information</CardTitle>
            <CardDescription>
              Config Server API endpoints
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <InfoCard label="API Base URL" value={process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'} />
              <InfoCard label="API Version" value="v1" />
              <InfoCard label="Health Endpoint" value="/health" />
              <InfoCard label="Metrics Endpoint" value="/metrics" />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function StatusRow({
  label,
  status,
  value,
}: {
  label: string
  status?: boolean
  value: string
}) {
  return (
    <div className="flex items-center justify-between py-2">
      <span className="text-sm text-gray-600">{label}</span>
      <div className="flex items-center gap-2">
        {status !== undefined && (
          status ? (
            <CheckCircle2 className="h-4 w-4 text-green-600" />
          ) : (
            <XCircle className="h-4 w-4 text-red-600" />
          )
        )}
        <Badge variant={status ? 'default' : status === false ? 'destructive' : 'secondary'}>
          {value}
        </Badge>
      </div>
    </div>
  )
}

function ExternalServiceLink({
  name,
  description,
  url,
}: {
  name: string
  description: string
  url: string
}) {
  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="flex items-center justify-between rounded-lg border p-4 transition-colors hover:bg-gray-50"
    >
      <div>
        <div className="font-medium">{name}</div>
        <div className="text-sm text-gray-500">{description}</div>
      </div>
      <ExternalLink className="h-4 w-4 text-gray-400" />
    </a>
  )
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border p-4">
      <div className="text-sm text-gray-500">{label}</div>
      <div className="mt-1 font-mono text-sm">{value}</div>
    </div>
  )
}
