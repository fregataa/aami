'use client'

import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { useGroups } from '@/lib/hooks/use-groups'
import type { AlertRule } from '@/types/api'

const formSchema = z.object({
  group_id: z.string().min(1, 'Group is required'),
  name: z.string().min(1, 'Name is required'),
  description: z.string(),
  severity: z.enum(['critical', 'warning', 'info']),
  query_template: z.string().min(1, 'Query template is required'),
  default_config: z.string(),
  config: z.string(),
  enabled: z.boolean(),
  priority: z.number().min(0),
})

type FormValues = z.infer<typeof formSchema>

interface AlertRuleFormProps {
  open: boolean
  onClose: () => void
  rule?: AlertRule | null
  onSubmit: (data: {
    group_id: string
    name: string
    description?: string
    severity: 'critical' | 'warning' | 'info'
    query_template: string
    default_config?: Record<string, unknown>
    config?: Record<string, unknown>
    enabled: boolean
    priority: number
  }) => Promise<void>
}

export function AlertRuleForm({ open, onClose, rule, onSubmit }: AlertRuleFormProps) {
  const { groups } = useGroups()

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      group_id: '',
      name: '',
      description: '',
      severity: 'warning',
      query_template: '',
      default_config: '{}',
      config: '{}',
      enabled: true,
      priority: 0,
    },
  })

  useEffect(() => {
    if (rule) {
      form.reset({
        group_id: rule.group_id,
        name: rule.name,
        description: rule.description || '',
        severity: rule.severity,
        query_template: rule.query_template,
        default_config: JSON.stringify(rule.default_config || {}, null, 2),
        config: JSON.stringify(rule.config || {}, null, 2),
        enabled: rule.enabled,
        priority: rule.priority,
      })
    } else {
      form.reset({
        group_id: '',
        name: '',
        description: '',
        severity: 'warning',
        query_template: '',
        default_config: '{}',
        config: '{}',
        enabled: true,
        priority: 0,
      })
    }
  }, [rule, form])

  const handleSubmit = async (data: FormValues) => {
    let defaultConfig: Record<string, unknown> = {}
    let config: Record<string, unknown> = {}

    try {
      defaultConfig = JSON.parse(data.default_config)
    } catch {
      form.setError('default_config', { message: 'Invalid JSON' })
      return
    }

    try {
      config = JSON.parse(data.config)
    } catch {
      form.setError('config', { message: 'Invalid JSON' })
      return
    }

    await onSubmit({
      group_id: data.group_id,
      name: data.name,
      description: data.description || undefined,
      severity: data.severity,
      query_template: data.query_template,
      default_config: defaultConfig,
      config: config,
      enabled: data.enabled,
      priority: data.priority,
    })
    form.reset()
  }

  const handleClose = () => {
    form.reset()
    onClose()
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{rule ? 'Edit Alert Rule' : 'Create Alert Rule'}</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input placeholder="HighCPUUsage" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="group_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Group</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select group" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {groups.map((group) => (
                          <SelectItem key={group.id} value={group.id}>
                            {group.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="severity"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Severity</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select severity" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="critical">Critical</SelectItem>
                        <SelectItem value="warning">Warning</SelectItem>
                        <SelectItem value="info">Info</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="priority"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Priority</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        placeholder="0"
                        {...field}
                        onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Input placeholder="Alert for high CPU usage" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="query_template"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Query Template</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="100 - (avg by(instance) (rate(node_cpu_seconds_total{mode='idle'}[5m])) * 100) > {{ .threshold }}"
                      className="font-mono text-sm"
                      rows={4}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Use Go template syntax for variables: {'{{ .variable }}'}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="default_config"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Default Config (JSON)</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='{"threshold": 80}'
                        className="font-mono text-sm"
                        rows={3}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="config"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Config Override (JSON)</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='{"threshold": 90}'
                        className="font-mono text-sm"
                        rows={3}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="enabled"
              render={({ field }) => (
                <FormItem className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <FormLabel>Enabled</FormLabel>
                    <FormDescription>
                      Enable or disable this alert rule
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                  </FormControl>
                </FormItem>
              )}
            />

            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={handleClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={form.formState.isSubmitting}>
                {form.formState.isSubmitting ? 'Saving...' : 'Save'}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
