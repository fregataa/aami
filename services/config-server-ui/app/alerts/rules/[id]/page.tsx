'use client'

import { use, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useAlertRule } from '@/lib/hooks/use-alert-rules'
import { alertRulesApi, type UpdateAlertRuleRequest } from '@/lib/api/alert-rules'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertRuleForm } from '@/components/alerts/rule-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { toast } from 'sonner'
import { ArrowLeft, Edit, Trash2 } from 'lucide-react'
import type { AlertRule } from '@/types/api'

export default function AlertRuleDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  const router = useRouter()
  const { rule, isLoading, mutate } = useAlertRule(id)

  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleUpdate = async (data: UpdateAlertRuleRequest) => {
    try {
      await alertRulesApi.update(id, data)
      toast.success('Rule updated successfully')
      mutate()
      setShowEdit(false)
    } catch (error) {
      toast.error('Failed to update rule')
    }
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      await alertRulesApi.delete(id)
      toast.success('Rule deleted successfully')
      router.push('/alerts/rules')
    } catch (error) {
      toast.error('Failed to delete rule')
      setIsDeleting(false)
    }
  }

  const handleToggleEnabled = async (enabled: boolean) => {
    try {
      await alertRulesApi.toggleEnabled(id, enabled)
      toast.success(`Rule ${enabled ? 'enabled' : 'disabled'}`)
      mutate()
    } catch (error) {
      toast.error('Failed to update rule')
    }
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64 w-full" />
      </div>
    )
  }

  if (!rule) {
    return (
      <div className="space-y-6">
        <Link href="/alerts/rules">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Rules
          </Button>
        </Link>
        <div className="text-center text-gray-500">Rule not found</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/alerts/rules">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <h1 className="text-2xl font-bold">{rule.name}</h1>
        <SeverityBadge severity={rule.severity} />
        <Badge variant={rule.enabled ? 'default' : 'secondary'}>
          {rule.enabled ? 'Enabled' : 'Disabled'}
        </Badge>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Rule Details</CardTitle>
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
            <div className="flex items-center justify-between">
              <div className="text-sm font-medium text-gray-500">Enabled</div>
              <Switch checked={rule.enabled} onCheckedChange={handleToggleEnabled} />
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Description</div>
              <div>{rule.description || '-'}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Group</div>
              {rule.group ? (
                <Link href={`/groups/${rule.group.id}`} className="text-blue-600 hover:underline">
                  {rule.group.name}
                </Link>
              ) : (
                <span className="text-gray-400">-</span>
              )}
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Priority</div>
              <div>{rule.priority}</div>
            </div>
            {rule.created_from_template_name && (
              <div>
                <div className="text-sm font-medium text-gray-500">Based on Template</div>
                <Link
                  href={`/alerts/templates/${rule.created_from_template_id}`}
                  className="text-blue-600 hover:underline"
                >
                  {rule.created_from_template_name}
                </Link>
              </div>
            )}
            <div>
              <div className="text-sm font-medium text-gray-500">Created</div>
              <div>{new Date(rule.created_at).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Updated</div>
              <div>{new Date(rule.updated_at).toLocaleString()}</div>
            </div>
          </CardContent>
        </Card>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Query Template</CardTitle>
            </CardHeader>
            <CardContent>
              <pre className="overflow-x-auto rounded bg-gray-100 p-3 text-sm">
                {rule.query_template}
              </pre>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Configuration</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <div className="text-sm font-medium text-gray-500">Default Config</div>
                <pre className="mt-1 rounded bg-gray-100 p-3 text-sm">
                  {JSON.stringify(rule.default_config, null, 2)}
                </pre>
              </div>
              <div>
                <div className="text-sm font-medium text-gray-500">Config Override</div>
                <pre className="mt-1 rounded bg-gray-100 p-3 text-sm">
                  {JSON.stringify(rule.config, null, 2)}
                </pre>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <AlertRuleForm
        open={showEdit}
        onClose={() => setShowEdit(false)}
        rule={rule}
        onSubmit={handleUpdate}
      />

      <DeleteDialog
        open={showDelete}
        onClose={() => setShowDelete(false)}
        onConfirm={handleDelete}
        title="Delete Rule"
        description={`Are you sure you want to delete "${rule.name}"? This action cannot be undone.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}

function SeverityBadge({ severity }: { severity: AlertRule['severity'] }) {
  const variants: Record<AlertRule['severity'], 'default' | 'secondary' | 'destructive'> = {
    critical: 'destructive',
    warning: 'default',
    info: 'secondary',
  }

  return <Badge variant={variants[severity]}>{severity}</Badge>
}
