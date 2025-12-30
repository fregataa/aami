'use client'

import { useState } from 'react'
import { useAlertRules } from '@/lib/hooks/use-alert-rules'
import { alertRulesApi, type CreateAlertRuleRequest, type UpdateAlertRuleRequest } from '@/lib/api/alert-rules'
import { Button } from '@/components/ui/button'
import { AlertRuleTable } from '@/components/alerts/rule-table'
import { AlertRuleForm } from '@/components/alerts/rule-form'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import type { AlertRule } from '@/types/api'

export default function AlertRulesPage() {
  const { rules, isLoading, mutate } = useAlertRules()
  const [showForm, setShowForm] = useState(false)
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null)
  const [deletingRule, setDeletingRule] = useState<AlertRule | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleCreate = async (data: CreateAlertRuleRequest) => {
    try {
      await alertRulesApi.create(data)
      toast.success('Alert rule created successfully')
      mutate()
      setShowForm(false)
    } catch (error) {
      toast.error('Failed to create alert rule')
    }
  }

  const handleUpdate = async (data: UpdateAlertRuleRequest) => {
    if (!editingRule) return
    try {
      await alertRulesApi.update(editingRule.id, data)
      toast.success('Alert rule updated successfully')
      mutate()
      setEditingRule(null)
    } catch (error) {
      toast.error('Failed to update alert rule')
    }
  }

  const handleDelete = async () => {
    if (!deletingRule) return
    setIsDeleting(true)
    try {
      await alertRulesApi.delete(deletingRule.id)
      toast.success('Alert rule deleted successfully')
      mutate()
      setDeletingRule(null)
    } catch (error) {
      toast.error('Failed to delete alert rule')
    } finally {
      setIsDeleting(false)
    }
  }

  const handleToggleEnabled = async (rule: AlertRule, enabled: boolean) => {
    try {
      await alertRulesApi.toggleEnabled(rule.id, enabled)
      toast.success(`Alert rule ${enabled ? 'enabled' : 'disabled'}`)
      mutate()
    } catch (error) {
      toast.error('Failed to update alert rule')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Alert Rules</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Rule
        </Button>
      </div>

      <AlertRuleTable
        rules={rules}
        isLoading={isLoading}
        onEdit={setEditingRule}
        onDelete={setDeletingRule}
        onToggleEnabled={handleToggleEnabled}
      />

      <AlertRuleForm
        open={showForm || !!editingRule}
        onClose={() => {
          setShowForm(false)
          setEditingRule(null)
        }}
        rule={editingRule}
        onSubmit={editingRule ? handleUpdate : handleCreate}
      />

      <DeleteDialog
        open={!!deletingRule}
        onClose={() => setDeletingRule(null)}
        onConfirm={handleDelete}
        title="Delete Alert Rule"
        description={`Are you sure you want to delete "${deletingRule?.name}"? This action cannot be undone.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}
