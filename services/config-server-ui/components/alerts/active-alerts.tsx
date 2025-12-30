'use client'

import { useActiveAlerts } from '@/lib/hooks/use-active-alerts'
import { Badge } from '@/components/ui/badge'
import { formatDistanceToNow } from 'date-fns'

export function ActiveAlerts() {
  const { alerts, total, isLoading } = useActiveAlerts()

  if (isLoading) {
    return <div className="text-gray-500">Loading alerts...</div>
  }

  if (!alerts.length) {
    return (
      <div className="flex h-32 items-center justify-center text-gray-500">
        No active alerts
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {alerts.slice(0, 5).map((alert) => (
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
      {total > 5 && (
        <div className="text-center text-sm text-gray-500">
          and {total - 5} more alerts
        </div>
      )}
    </div>
  )
}

function SeverityBadge({ severity }: { severity?: string }) {
  const variant = severity === 'critical' ? 'destructive'
    : severity === 'warning' ? 'secondary'
    : 'outline'

  return (
    <Badge variant={variant}>
      {severity?.toUpperCase() || 'UNKNOWN'}
    </Badge>
  )
}
