'use client'

import { useState } from 'react'
import { useScriptTemplates } from '@/lib/hooks/use-script-templates'
import { scriptTemplatesApi, type CreateScriptTemplateRequest, type UpdateScriptTemplateRequest } from '@/lib/api/script-templates'
import { Button } from '@/components/ui/button'
import { ScriptTemplateTable } from '@/components/scripts/template-table'
import { ScriptTemplateForm } from '@/components/scripts/template-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import type { ScriptTemplate } from '@/types/api'

export default function ScriptTemplatesPage() {
  const { templates, isLoading, mutate } = useScriptTemplates()
  const [showForm, setShowForm] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<ScriptTemplate | null>(null)
  const [deletingTemplate, setDeletingTemplate] = useState<ScriptTemplate | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleCreate = async (data: CreateScriptTemplateRequest) => {
    try {
      await scriptTemplatesApi.create(data)
      toast.success('Script template created successfully')
      mutate()
      setShowForm(false)
    } catch (error) {
      toast.error('Failed to create script template')
    }
  }

  const handleUpdate = async (data: UpdateScriptTemplateRequest) => {
    if (!editingTemplate) return
    try {
      await scriptTemplatesApi.update(editingTemplate.id, data)
      toast.success('Script template updated successfully')
      mutate()
      setEditingTemplate(null)
    } catch (error) {
      toast.error('Failed to update script template')
    }
  }

  const handleDelete = async () => {
    if (!deletingTemplate) return
    setIsDeleting(true)
    try {
      await scriptTemplatesApi.delete(deletingTemplate.id)
      toast.success('Script template deleted successfully')
      mutate()
      setDeletingTemplate(null)
    } catch (error) {
      toast.error('Failed to delete script template')
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Script Templates</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Template
        </Button>
      </div>

      <ScriptTemplateTable
        templates={templates}
        isLoading={isLoading}
        onEdit={setEditingTemplate}
        onDelete={setDeletingTemplate}
      />

      <ScriptTemplateForm
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
        title="Delete Script Template"
        description={`Are you sure you want to delete "${deletingTemplate?.name}"? This action cannot be undone.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}
