'use client'

import { use, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useAlertTemplate } from '@/lib/hooks/use-alert-templates'
import { useAlertRulesByTemplate } from '@/lib/hooks/use-alert-rules'
import { alertTemplatesApi, type UpdateAlertTemplateRequest } from '@/lib/api/alert-templates'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertTemplateForm } from '@/components/alerts/template-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { toast } from 'sonner'
import { ArrowLeft, Edit, Trash2 } from 'lucide-react'
import type { AlertTemplate } from '@/types/api'

export default function AlertTemplateDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  const router = useRouter()
  const { template, isLoading, mutate } = useAlertTemplate(id)
  const { rules } = useAlertRulesByTemplate(id)

  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleUpdate = async (data: UpdateAlertTemplateRequest) => {
    try {
      await alertTemplatesApi.update(id, data)
      toast.success('Template updated successfully')
      mutate()
      setShowEdit(false)
    } catch (error) {
      toast.error('Failed to update template')
    }
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      await alertTemplatesApi.delete(id)
      toast.success('Template deleted successfully')
      router.push('/alerts/templates')
    } catch (error) {
      toast.error('Failed to delete template')
      setIsDeleting(false)
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

  if (!template) {
    return (
      <div className="space-y-6">
        <Link href="/alerts/templates">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Templates
          </Button>
        </Link>
        <div className="text-center text-gray-500">Template not found</div>
      </div>
    )
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
              <pre className="mt-1 overflow-x-auto rounded bg-gray-100 p-3 text-sm">
                {template.query_template}
              </pre>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Default Config</div>
              <pre className="mt-1 rounded bg-gray-100 p-3 text-sm">
                {JSON.stringify(template.default_config, null, 2)}
              </pre>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Created</div>
              <div>{new Date(template.created_at).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Updated</div>
              <div>{new Date(template.updated_at).toLocaleString()}</div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Rules Using This Template ({rules.length})</CardTitle>
          </CardHeader>
          <CardContent>
            {rules.length > 0 ? (
              <div className="space-y-2">
                {rules.map((rule) => (
                  <Link
                    key={rule.id}
                    href={`/alerts/rules/${rule.id}`}
                    className="block rounded-lg border p-3 hover:bg-gray-50"
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">{rule.name}</div>
                        <div className="text-sm text-gray-500">
                          Group: {rule.group?.name ?? 'Unknown'}
                        </div>
                      </div>
                      <Badge variant={rule.enabled ? 'default' : 'secondary'}>
                        {rule.enabled ? 'Enabled' : 'Disabled'}
                      </Badge>
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
        description={`Are you sure you want to delete "${template.name}"? This action cannot be undone.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}

function SeverityBadge({ severity }: { severity: AlertTemplate['severity'] }) {
  const variants: Record<AlertTemplate['severity'], 'default' | 'secondary' | 'destructive'> = {
    critical: 'destructive',
    warning: 'default',
    info: 'secondary',
  }

  return <Badge variant={variants[severity]}>{severity}</Badge>
}
