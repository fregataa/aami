'use client'

import { useState } from 'react'
import { useScriptPolicies } from '@/lib/hooks/use-script-policies'
import { scriptPoliciesApi, type CreateScriptPolicyRequest, type UpdateScriptPolicyRequest } from '@/lib/api/script-policies'
import { Button } from '@/components/ui/button'
import { ScriptPolicyTable } from '@/components/scripts/policy-table'
import { ScriptPolicyForm } from '@/components/scripts/policy-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import type { ScriptPolicy } from '@/types/api'

export default function ScriptPoliciesPage() {
  const { policies, isLoading, mutate } = useScriptPolicies()
  const [showForm, setShowForm] = useState(false)
  const [editingPolicy, setEditingPolicy] = useState<ScriptPolicy | null>(null)
  const [deletingPolicy, setDeletingPolicy] = useState<ScriptPolicy | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleCreate = async (data: CreateScriptPolicyRequest) => {
    try {
      await scriptPoliciesApi.create(data)
      toast.success('Script policy created successfully')
      mutate()
      setShowForm(false)
    } catch (error) {
      toast.error('Failed to create script policy')
    }
  }

  const handleUpdate = async (data: UpdateScriptPolicyRequest) => {
    if (!editingPolicy) return
    try {
      await scriptPoliciesApi.update(editingPolicy.id, data)
      toast.success('Script policy updated successfully')
      mutate()
      setEditingPolicy(null)
    } catch (error) {
      toast.error('Failed to update script policy')
    }
  }

  const handleDelete = async () => {
    if (!deletingPolicy) return
    setIsDeleting(true)
    try {
      await scriptPoliciesApi.delete(deletingPolicy.id)
      toast.success('Script policy deleted successfully')
      mutate()
      setDeletingPolicy(null)
    } catch (error) {
      toast.error('Failed to delete script policy')
    } finally {
      setIsDeleting(false)
    }
  }

  const handleToggleEnabled = async (policy: ScriptPolicy, enabled: boolean) => {
    try {
      await scriptPoliciesApi.update(policy.id, { enabled })
      toast.success(`Script policy ${enabled ? 'enabled' : 'disabled'}`)
      mutate()
    } catch (error) {
      toast.error('Failed to update script policy')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Script Policies</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Policy
        </Button>
      </div>

      <ScriptPolicyTable
        policies={policies}
        isLoading={isLoading}
        onEdit={setEditingPolicy}
        onDelete={setDeletingPolicy}
        onToggleEnabled={handleToggleEnabled}
      />

      <ScriptPolicyForm
        open={showForm || !!editingPolicy}
        onClose={() => {
          setShowForm(false)
          setEditingPolicy(null)
        }}
        policy={editingPolicy}
        onSubmit={editingPolicy ? handleUpdate : handleCreate}
      />

      <DeleteDialog
        open={!!deletingPolicy}
        onClose={() => setDeletingPolicy(null)}
        onConfirm={handleDelete}
        title="Delete Script Policy"
        description="Are you sure you want to delete this policy? This action cannot be undone."
        isDeleting={isDeleting}
      />
    </div>
  )
}
