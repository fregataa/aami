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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import type { AlertTemplate } from '@/types/api'

const formSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string(),
  severity: z.enum(['critical', 'warning', 'info']),
  query_template: z.string().min(1, 'Query template is required'),
  default_config: z.string(),
})

type FormValues = z.infer<typeof formSchema>

interface AlertTemplateFormProps {
  open: boolean
  onClose: () => void
  template?: AlertTemplate | null
  onSubmit: (data: {
    name: string
    description?: string
    severity: 'critical' | 'warning' | 'info'
    query_template: string
    default_config?: Record<string, unknown>
  }) => Promise<void>
}

export function AlertTemplateForm({ open, onClose, template, onSubmit }: AlertTemplateFormProps) {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      description: '',
      severity: 'warning',
      query_template: '',
      default_config: '{}',
    },
  })

  useEffect(() => {
    if (template) {
      form.reset({
        name: template.name,
        description: template.description || '',
        severity: template.severity,
        query_template: template.query_template,
        default_config: JSON.stringify(template.default_config || {}, null, 2),
      })
    } else {
      form.reset({
        name: '',
        description: '',
        severity: 'warning',
        query_template: '',
        default_config: '{}',
      })
    }
  }, [template, form])

  const handleSubmit = async (data: FormValues) => {
    let defaultConfig: Record<string, unknown> = {}
    try {
      defaultConfig = JSON.parse(data.default_config)
    } catch {
      form.setError('default_config', { message: 'Invalid JSON' })
      return
    }

    await onSubmit({
      name: data.name,
      description: data.description || undefined,
      severity: data.severity,
      query_template: data.query_template,
      default_config: defaultConfig,
    })
    form.reset()
  }

  const handleClose = () => {
    form.reset()
    onClose()
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{template ? 'Edit Alert Template' : 'Create Alert Template'}</DialogTitle>
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
                  <FormDescription>
                    Default values for template variables
                  </FormDescription>
                  <FormMessage />
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
