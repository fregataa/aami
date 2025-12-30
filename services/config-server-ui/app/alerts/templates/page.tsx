'use client'

import { useState } from 'react'
import { useAlertTemplates } from '@/lib/hooks/use-alert-templates'
import { alertTemplatesApi, type CreateAlertTemplateRequest, type UpdateAlertTemplateRequest } from '@/lib/api/alert-templates'
import { Button } from '@/components/ui/button'
import { AlertTemplateTable } from '@/components/alerts/template-table'
import { AlertTemplateForm } from '@/components/alerts/template-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import type { AlertTemplate } from '@/types/api'

export default function AlertTemplatesPage() {
  const { templates, isLoading, mutate } = useAlertTemplates()
  const [showForm, setShowForm] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<AlertTemplate | null>(null)
  const [deletingTemplate, setDeletingTemplate] = useState<AlertTemplate | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleCreate = async (data: CreateAlertTemplateRequest) => {
    try {
      await alertTemplatesApi.create(data)
      toast.success('Alert template created successfully')
      mutate()
      setShowForm(false)
    } catch (error) {
      toast.error('Failed to create alert template')
    }
  }

  const handleUpdate = async (data: UpdateAlertTemplateRequest) => {
    if (!editingTemplate) return
    try {
      await alertTemplatesApi.update(editingTemplate.id, data)
      toast.success('Alert template updated successfully')
      mutate()
      setEditingTemplate(null)
    } catch (error) {
      toast.error('Failed to update alert template')
    }
  }

  const handleDelete = async () => {
    if (!deletingTemplate) return
    setIsDeleting(true)
    try {
      await alertTemplatesApi.delete(deletingTemplate.id)
      toast.success('Alert template deleted successfully')
      mutate()
      setDeletingTemplate(null)
    } catch (error) {
      toast.error('Failed to delete alert template')
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Alert Templates</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Template
        </Button>
      </div>

      <AlertTemplateTable
        templates={templates}
        isLoading={isLoading}
        onEdit={setEditingTemplate}
        onDelete={setDeletingTemplate}
      />

      <AlertTemplateForm
        open={showForm || !!editingTemplate}
        onClose={() => {
          setShowForm(false)
          setEditingTemplate(null)
        }}
        template={editingTemplate}
        onSubmit={editingTemplate ? handleUpdate : handleCreate}
      />

      <DeleteDialog
        open={!!deletingTemplate}
        onClose={() => setDeletingTemplate(null)}
        onConfirm={handleDelete}
        title="Delete Alert Template"
        description={`Are you sure you want to delete "${deletingTemplate?.name}"? This action cannot be undone.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}
