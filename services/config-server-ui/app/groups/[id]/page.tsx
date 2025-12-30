'use client'

import { use, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import useSWR from 'swr'
import { useGroup } from '@/lib/hooks/use-groups'
import { groupsApi, type UpdateGroupRequest } from '@/lib/api/groups'
import { targetsApi } from '@/lib/api/targets'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { GroupForm } from '@/components/groups/group-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { toast } from 'sonner'
import { ArrowLeft, Edit, Trash2 } from 'lucide-react'
import type { Target } from '@/types/api'

export default function GroupDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  const router = useRouter()
  const { group, isLoading, mutate } = useGroup(id)

  // Fetch targets in this group
  const { data: targets } = useSWR<Target[]>(
    id ? `/api/v1/targets/group/${id}` : null,
    () => targetsApi.getByGroup(id)
  )

  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleUpdate = async (data: UpdateGroupRequest) => {
    try {
      await groupsApi.update(id, data)
      toast.success('Group updated successfully')
      mutate()
      setShowEdit(false)
    } catch (error) {
      toast.error('Failed to update group')
    }
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      await groupsApi.delete(id)
      toast.success('Group deleted successfully')
      router.push('/groups')
    } catch (error) {
      toast.error('Failed to delete group')
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

  if (!group) {
    return (
      <div className="space-y-6">
        <Link href="/groups">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Groups
          </Button>
        </Link>
        <div className="text-center text-gray-500">Group not found</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/groups">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <h1 className="text-2xl font-bold">{group.name}</h1>
        {group.is_default_own && <Badge variant="secondary">Default</Badge>}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Group Details</CardTitle>
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
              <div className="text-sm font-medium text-gray-500">Name</div>
              <div>{group.name}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Description</div>
              <div>{group.description || '-'}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Priority</div>
              <div>{group.priority}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Created</div>
              <div>{new Date(group.created_at).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Updated</div>
              <div>{new Date(group.updated_at).toLocaleString()}</div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Targets in Group ({targets?.length ?? 0})</CardTitle>
          </CardHeader>
          <CardContent>
            {targets && targets.length > 0 ? (
              <div className="space-y-2">
                {targets.map((target) => (
                  <Link
                    key={target.id}
                    href={`/targets/${target.id}`}
                    className="block rounded-lg border p-3 hover:bg-gray-50"
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">{target.hostname}</div>
                        <div className="text-sm text-gray-500">{target.ip_address}</div>
                      </div>
                      <StatusBadge status={target.status} />
                    </div>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="text-gray-500">No targets in this group</div>
            )}
          </CardContent>
        </Card>

        {group.metadata && Object.keys(group.metadata).length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>Metadata</CardTitle>
            </CardHeader>
            <CardContent>
              <pre className="rounded bg-gray-100 p-3 text-sm">
                {JSON.stringify(group.metadata, null, 2)}
              </pre>
            </CardContent>
          </Card>
        )}
      </div>

      <GroupForm
        open={showEdit}
        onClose={() => setShowEdit(false)}
        group={group}
        onSubmit={handleUpdate}
      />

      <DeleteDialog
        open={showDelete}
        onClose={() => setShowDelete(false)}
        onConfirm={handleDelete}
        title="Delete Group"
        description={`Are you sure you want to delete "${group.name}"? This action cannot be undone.`}
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
