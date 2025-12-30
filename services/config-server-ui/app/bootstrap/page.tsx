'use client'

import { useState } from 'react'
import { useBootstrapTokens } from '@/lib/hooks/use-bootstrap-tokens'
import {
  bootstrapTokensApi,
  type CreateBootstrapTokenRequest,
  type UpdateBootstrapTokenRequest,
} from '@/lib/api/bootstrap-tokens'
import { Button } from '@/components/ui/button'
import { BootstrapTokenTable } from '@/components/bootstrap/bootstrap-token-table'
import { BootstrapTokenForm } from '@/components/bootstrap/bootstrap-token-form'
import { CommandGenerator } from '@/components/bootstrap/command-generator'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import type { BootstrapToken } from '@/types/api'

export default function BootstrapPage() {
  const { tokens, isLoading, mutate } = useBootstrapTokens()
  const [showForm, setShowForm] = useState(false)
  const [editingToken, setEditingToken] = useState<BootstrapToken | null>(null)
  const [deletingToken, setDeletingToken] = useState<BootstrapToken | null>(null)
  const [selectedToken, setSelectedToken] = useState<BootstrapToken | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)
  const [newlyCreatedToken, setNewlyCreatedToken] = useState<BootstrapToken | null>(null)

  const handleCreate = async (data: {
    name: string
    description?: string
    group_id: string
    expires_in_days?: number
    max_uses?: number
  }) => {
    try {
      const request: CreateBootstrapTokenRequest = {
        name: data.name,
        description: data.description,
        group_id: data.group_id,
        max_uses: data.max_uses || 0,
      }

      if (data.expires_in_days && data.expires_in_days > 0) {
        const expiresAt = new Date()
        expiresAt.setDate(expiresAt.getDate() + data.expires_in_days)
        request.expires_at = expiresAt.toISOString()
      }

      const createdToken = await bootstrapTokensApi.create(request)
      toast.success('Bootstrap token created successfully')

      // Store the newly created token with its value for command generation
      setNewlyCreatedToken(createdToken)
      setSelectedToken(createdToken)

      mutate()
      setShowForm(false)
    } catch {
      toast.error('Failed to create bootstrap token')
    }
  }

  const handleUpdate = async (data: {
    name: string
    description?: string
    group_id: string
    max_uses?: number
  }) => {
    if (!editingToken) return
    try {
      const request: UpdateBootstrapTokenRequest = {
        name: data.name,
        description: data.description,
        group_id: data.group_id,
        max_uses: data.max_uses,
      }

      await bootstrapTokensApi.update(editingToken.id, request)
      toast.success('Bootstrap token updated successfully')
      mutate()
      setEditingToken(null)
    } catch {
      toast.error('Failed to update bootstrap token')
    }
  }

  const handleDelete = async () => {
    if (!deletingToken) return
    setIsDeleting(true)
    try {
      await bootstrapTokensApi.delete(deletingToken.id)
      toast.success('Bootstrap token deleted successfully')

      // Clear selection if the deleted token was selected
      if (selectedToken?.id === deletingToken.id) {
        setSelectedToken(null)
      }

      mutate()
      setDeletingToken(null)
    } catch {
      toast.error('Failed to delete bootstrap token')
    } finally {
      setIsDeleting(false)
    }
  }

  const handleSelectToken = (token: BootstrapToken) => {
    // If this is the newly created token, use it (it has the token value)
    if (newlyCreatedToken?.id === token.id) {
      setSelectedToken(newlyCreatedToken)
    } else {
      setSelectedToken(token)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Bootstrap Tokens</h1>
          <p className="text-gray-500">
            Manage tokens for automated node registration
          </p>
        </div>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Token
        </Button>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="lg:col-span-2">
          <BootstrapTokenTable
            tokens={tokens}
            isLoading={isLoading}
            onEdit={setEditingToken}
            onDelete={setDeletingToken}
            onSelect={handleSelectToken}
          />
        </div>

        <div className="lg:col-span-2">
          <CommandGenerator token={selectedToken} />
        </div>
      </div>

      <BootstrapTokenForm
        open={showForm || !!editingToken}
        onClose={() => {
          setShowForm(false)
          setEditingToken(null)
        }}
        token={editingToken}
        onSubmit={editingToken ? handleUpdate : handleCreate}
      />

      <DeleteDialog
        open={!!deletingToken}
        onClose={() => setDeletingToken(null)}
        onConfirm={handleDelete}
        title="Delete Bootstrap Token"
        description={`Are you sure you want to delete "${deletingToken?.name}"? This action cannot be undone and any pending node registrations using this token will fail.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}
