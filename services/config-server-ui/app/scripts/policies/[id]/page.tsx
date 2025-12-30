'use client'

import { use, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useScriptPolicy } from '@/lib/hooks/use-script-policies'
import { scriptPoliciesApi, type UpdateScriptPolicyRequest } from '@/lib/api/script-policies'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { ScriptPolicyForm } from '@/components/scripts/policy-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { toast } from 'sonner'
import { ArrowLeft, Edit, Trash2 } from 'lucide-react'

export default function ScriptPolicyDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  const router = useRouter()
  const { policy, isLoading, mutate } = useScriptPolicy(id)

  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleUpdate = async (data: UpdateScriptPolicyRequest) => {
    try {
      await scriptPoliciesApi.update(id, data)
      toast.success('Policy updated successfully')
      mutate()
      setShowEdit(false)
    } catch (error) {
      toast.error('Failed to update policy')
    }
  }

  const handleToggleEnabled = async (enabled: boolean) => {
    try {
      await scriptPoliciesApi.update(id, { enabled })
      toast.success(`Policy ${enabled ? 'enabled' : 'disabled'}`)
      mutate()
    } catch (error) {
      toast.error('Failed to update policy')
    }
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      await scriptPoliciesApi.delete(id)
      toast.success('Policy deleted successfully')
      router.push('/scripts/policies')
    } catch (error) {
      toast.error('Failed to delete policy')
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

  if (!policy) {
    return (
      <div className="space-y-6">
        <Link href="/scripts/policies">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Policies
          </Button>
        </Link>
        <div className="text-center text-gray-500">Policy not found</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/scripts/policies">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <h1 className="text-2xl font-bold">
          {policy.template?.name ?? 'Unknown Template'}
        </h1>
        <Badge variant={policy.enabled ? 'default' : 'secondary'}>
          {policy.enabled ? 'Enabled' : 'Disabled'}
        </Badge>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Policy Details</CardTitle>
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
              <Switch checked={policy.enabled} onCheckedChange={handleToggleEnabled} />
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Template</div>
              {policy.template ? (
                <Link
                  href={`/scripts/templates/${policy.template.id}`}
                  className="text-blue-600 hover:underline"
                >
                  {policy.template.name}
                </Link>
              ) : (
                <span className="text-gray-400">Unknown</span>
              )}
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Group</div>
              {policy.group ? (
                <Link
                  href={`/groups/${policy.group.id}`}
                  className="text-blue-600 hover:underline"
                >
                  {policy.group.name}
                </Link>
              ) : (
                <Badge variant="secondary">Global (all targets)</Badge>
              )}
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Priority</div>
              <div>{policy.priority}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Created</div>
              <div>{new Date(policy.created_at).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Updated</div>
              <div>{new Date(policy.updated_at).toLocaleString()}</div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Configuration Override</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="overflow-x-auto rounded bg-gray-100 p-3 text-sm">
              {JSON.stringify(policy.config, null, 2)}
            </pre>
          </CardContent>
        </Card>
      </div>

      <ScriptPolicyForm
        open={showEdit}
        onClose={() => setShowEdit(false)}
        policy={policy}
        onSubmit={handleUpdate}
      />

      <DeleteDialog
        open={showDelete}
        onClose={() => setShowDelete(false)}
        onConfirm={handleDelete}
        title="Delete Policy"
        description="Are you sure you want to delete this policy? This action cannot be undone."
        isDeleting={isDeleting}
      />
    </div>
  )
}
