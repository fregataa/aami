'use client'

import { useState } from 'react'
import { useExporters } from '@/lib/hooks/use-exporters'
import { exportersApi, type CreateExporterRequest, type UpdateExporterRequest } from '@/lib/api/exporters'
import { Button } from '@/components/ui/button'
import { ExporterTable } from '@/components/exporters/exporter-table'
import { ExporterForm } from '@/components/exporters/exporter-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import type { Exporter } from '@/types/api'

export default function ExportersPage() {
  const { exporters, isLoading, mutate } = useExporters()
  const [showForm, setShowForm] = useState(false)
  const [editingExporter, setEditingExporter] = useState<Exporter | null>(null)
  const [deletingExporter, setDeletingExporter] = useState<Exporter | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleCreate = async (data: CreateExporterRequest) => {
    try {
      await exportersApi.create({ ...data, enabled: true })
      toast.success('Exporter created successfully')
      mutate()
      setShowForm(false)
    } catch (error) {
      toast.error('Failed to create exporter')
    }
  }

  const handleUpdate = async (data: UpdateExporterRequest) => {
    if (!editingExporter) return
    try {
      await exportersApi.update(editingExporter.id, data)
      toast.success('Exporter updated successfully')
      mutate()
      setEditingExporter(null)
    } catch (error) {
      toast.error('Failed to update exporter')
    }
  }

  const handleToggle = async (exporter: Exporter) => {
    try {
      await exportersApi.update(exporter.id, { enabled: !exporter.enabled })
      toast.success(`Exporter ${exporter.enabled ? 'disabled' : 'enabled'} successfully`)
      mutate()
    } catch (error) {
      toast.error('Failed to toggle exporter')
    }
  }

  const handleDelete = async () => {
    if (!deletingExporter) return
    setIsDeleting(true)
    try {
      await exportersApi.delete(deletingExporter.id)
      toast.success('Exporter deleted successfully')
      mutate()
      setDeletingExporter(null)
    } catch (error) {
      toast.error('Failed to delete exporter')
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Exporters</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Exporter
        </Button>
      </div>

      <ExporterTable
        exporters={exporters}
        isLoading={isLoading}
        onEdit={setEditingExporter}
        onDelete={setDeletingExporter}
        onToggle={handleToggle}
      />

      <ExporterForm
        open={showForm || !!editingExporter}
        onClose={() => {
          setShowForm(false)
          setEditingExporter(null)
        }}
        exporter={editingExporter}
        onSubmit={editingExporter ? handleUpdate : handleCreate}
      />

      <DeleteDialog
        open={!!deletingExporter}
        onClose={() => setDeletingExporter(null)}
        onConfirm={handleDelete}
        title="Delete Exporter"
        description={`Are you sure you want to delete this ${deletingExporter?.type} exporter? This action cannot be undone.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}
