'use client'

import { useState } from 'react'
import { useGroups } from '@/lib/hooks/use-groups'
import { groupsApi, type CreateGroupRequest, type UpdateGroupRequest } from '@/lib/api/groups'
import { Button } from '@/components/ui/button'
import { GroupTable } from '@/components/groups/group-table'
import { GroupForm } from '@/components/groups/group-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import type { Group } from '@/types/api'

export default function GroupsPage() {
  const { groups, isLoading, mutate } = useGroups()
  const [showForm, setShowForm] = useState(false)
  const [editingGroup, setEditingGroup] = useState<Group | null>(null)
  const [deletingGroup, setDeletingGroup] = useState<Group | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleCreate = async (data: CreateGroupRequest) => {
    try {
      await groupsApi.create(data)
      toast.success('Group created successfully')
      mutate()
      setShowForm(false)
    } catch (error) {
      toast.error('Failed to create group')
    }
  }

  const handleUpdate = async (data: UpdateGroupRequest) => {
    if (!editingGroup) return
    try {
      await groupsApi.update(editingGroup.id, data)
      toast.success('Group updated successfully')
      mutate()
      setEditingGroup(null)
    } catch (error) {
      toast.error('Failed to update group')
    }
  }

  const handleDelete = async () => {
    if (!deletingGroup) return
    setIsDeleting(true)
    try {
      await groupsApi.delete(deletingGroup.id)
      toast.success('Group deleted successfully')
      mutate()
      setDeletingGroup(null)
    } catch (error) {
      toast.error('Failed to delete group')
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Groups</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Group
        </Button>
      </div>

      <GroupTable
        groups={groups}
        isLoading={isLoading}
        onEdit={setEditingGroup}
        onDelete={setDeletingGroup}
      />

      <GroupForm
        open={showForm || !!editingGroup}
        onClose={() => {
          setShowForm(false)
          setEditingGroup(null)
        }}
        group={editingGroup}
        onSubmit={editingGroup ? handleUpdate : handleCreate}
      />

      <DeleteDialog
        open={!!deletingGroup}
        onClose={() => setDeletingGroup(null)}
        onConfirm={handleDelete}
        title="Delete Group"
        description={`Are you sure you want to delete "${deletingGroup?.name}"? This action cannot be undone.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}
