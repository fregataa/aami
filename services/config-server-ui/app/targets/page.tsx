'use client'

import { useState } from 'react'
import { useTargets } from '@/lib/hooks/use-targets'
import { targetsApi, type CreateTargetRequest, type UpdateTargetRequest } from '@/lib/api/targets'
import { Button } from '@/components/ui/button'
import { TargetTable } from '@/components/targets/target-table'
import { TargetForm } from '@/components/targets/target-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import type { Target } from '@/types/api'

export default function TargetsPage() {
  const { targets, isLoading, mutate } = useTargets()
  const [showForm, setShowForm] = useState(false)
  const [editingTarget, setEditingTarget] = useState<Target | null>(null)
  const [deletingTarget, setDeletingTarget] = useState<Target | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleCreate = async (data: CreateTargetRequest) => {
    try {
      await targetsApi.create(data)
      toast.success('Target created successfully')
      mutate()
      setShowForm(false)
    } catch (error) {
      toast.error('Failed to create target')
    }
  }

  const handleUpdate = async (data: UpdateTargetRequest) => {
    if (!editingTarget) return
    try {
      await targetsApi.update(editingTarget.id, data)
      toast.success('Target updated successfully')
      mutate()
      setEditingTarget(null)
    } catch (error) {
      toast.error('Failed to update target')
    }
  }

  const handleDelete = async () => {
    if (!deletingTarget) return
    setIsDeleting(true)
    try {
      await targetsApi.delete(deletingTarget.id)
      toast.success('Target deleted successfully')
      mutate()
      setDeletingTarget(null)
    } catch (error) {
      toast.error('Failed to delete target')
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Targets</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Target
        </Button>
      </div>

      <TargetTable
        targets={targets}
        isLoading={isLoading}
        onEdit={setEditingTarget}
        onDelete={setDeletingTarget}
      />

      <TargetForm
        open={showForm || !!editingTarget}
        onClose={() => {
          setShowForm(false)
          setEditingTarget(null)
        }}
        target={editingTarget}
        onSubmit={editingTarget ? handleUpdate : handleCreate}
      />

      <DeleteDialog
        open={!!deletingTarget}
        onClose={() => setDeletingTarget(null)}
        onConfirm={handleDelete}
        title="Delete Target"
        description={`Are you sure you want to delete "${deletingTarget?.hostname}"? This action cannot be undone.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}
