'use client'

import { use, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useTarget } from '@/lib/hooks/use-targets'
import { targetsApi, type UpdateTargetRequest } from '@/lib/api/targets'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { TargetForm } from '@/components/targets/target-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { toast } from 'sonner'
import { ArrowLeft, Edit, Trash2 } from 'lucide-react'
import type { Target } from '@/types/api'

export default function TargetDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  const router = useRouter()
  const { target, isLoading, mutate } = useTarget(id)

  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleUpdate = async (data: UpdateTargetRequest) => {
    try {
      await targetsApi.update(id, data)
      toast.success('Target updated successfully')
      mutate()
      setShowEdit(false)
    } catch (error) {
      toast.error('Failed to update target')
    }
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      await targetsApi.delete(id)
      toast.success('Target deleted successfully')
      router.push('/targets')
    } catch (error) {
      toast.error('Failed to delete target')
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

  if (!target) {
    return (
      <div className="space-y-6">
        <Link href="/targets">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Targets
          </Button>
        </Link>
        <div className="text-center text-gray-500">Target not found</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/targets">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <h1 className="text-2xl font-bold">{target.hostname}</h1>
        <StatusBadge status={target.status} />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Target Details</CardTitle>
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
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-sm font-medium text-gray-500">Hostname</div>
                <div className="font-mono">{target.hostname}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-gray-500">IP Address</div>
                <div className="font-mono">{target.ip_address}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-gray-500">Port</div>
                <div>{target.port}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-gray-500">Status</div>
                <StatusBadge status={target.status} />
              </div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Created</div>
              <div>{new Date(target.created_at).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Updated</div>
              <div>{new Date(target.updated_at).toLocaleString()}</div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Groups</CardTitle>
          </CardHeader>
          <CardContent>
            {target.groups && target.groups.length > 0 ? (
              <div className="flex flex-wrap gap-2">
                {target.groups.map((group) => (
                  <Link key={group.id} href={`/groups/${group.id}`}>
                    <Badge variant="outline" className="cursor-pointer hover:bg-gray-100">
                      {group.name}
                    </Badge>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="text-gray-500">No groups assigned</div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Labels</CardTitle>
          </CardHeader>
          <CardContent>
            {target.labels && Object.keys(target.labels).length > 0 ? (
              <div className="space-y-2">
                {Object.entries(target.labels).map(([key, value]) => (
                  <div key={key} className="flex items-center gap-2">
                    <Badge variant="secondary">{key}</Badge>
                    <span className="text-sm">{value}</span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-gray-500">No labels</div>
            )}
          </CardContent>
        </Card>
      </div>

      <TargetForm
        open={showEdit}
        onClose={() => setShowEdit(false)}
        target={target}
        onSubmit={handleUpdate}
      />

      <DeleteDialog
        open={showDelete}
        onClose={() => setShowDelete(false)}
        onConfirm={handleDelete}
        title="Delete Target"
        description={`Are you sure you want to delete "${target.hostname}"? This action cannot be undone.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}

function StatusBadge({ status }: { status: Target['status'] }) {
  const variants: Record<Target['status'], 'default' | 'secondary' | 'destructive' | 'outline'> = {
    active: 'default',
    inactive: 'secondary',
    down: 'destructive',
  }

  return <Badge variant={variants[status] || 'outline'}>{status}</Badge>
}
