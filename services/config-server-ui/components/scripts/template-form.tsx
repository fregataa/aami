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
import type { ScriptTemplate } from '@/types/api'

const formSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string(),
  script_type: z.string().min(1, 'Script type is required'),
  script_content: z.string().min(1, 'Script content is required'),
  config_schema: z.string(),
  enabled: z.boolean(),
})

type FormValues = z.infer<typeof formSchema>

interface ScriptTemplateFormProps {
  open: boolean
  onClose: () => void
  template?: ScriptTemplate | null
  onSubmit: (data: {
    name: string
    description?: string
    script_type: string
    script_content: string
    config_schema?: Record<string, unknown>
    enabled: boolean
  }) => Promise<void>
}

export function ScriptTemplateForm({ open, onClose, template, onSubmit }: ScriptTemplateFormProps) {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      description: '',
      script_type: 'bash',
      script_content: '#!/bin/bash\n\n',
      config_schema: '{}',
      enabled: true,
    },
  })

  useEffect(() => {
    if (template) {
      form.reset({
        name: template.name,
        description: template.description || '',
        script_type: template.script_type,
        script_content: template.script_content,
        config_schema: JSON.stringify(template.config_schema || {}, null, 2),
        enabled: template.enabled,
      })
    } else {
      form.reset({
        name: '',
        description: '',
        script_type: 'bash',
        script_content: '#!/bin/bash\n\n',
        config_schema: '{}',
        enabled: true,
      })
    }
  }, [template, form])

  const handleSubmit = async (data: FormValues) => {
    let configSchema: Record<string, unknown> = {}
    try {
      configSchema = JSON.parse(data.config_schema)
    } catch {
      form.setError('config_schema', { message: 'Invalid JSON' })
      return
    }

    await onSubmit({
      name: data.name,
      description: data.description || undefined,
      script_type: data.script_type,
      script_content: data.script_content,
      config_schema: configSchema,
      enabled: data.enabled,
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
          <DialogTitle>{template ? 'Edit Script Template' : 'Create Script Template'}</DialogTitle>
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
                      <Input placeholder="system-health-check" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="script_type"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Type</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select type" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="bash">Bash</SelectItem>
                        <SelectItem value="python">Python</SelectItem>
                        <SelectItem value="powershell">PowerShell</SelectItem>
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
                    <Input placeholder="Script to check system health" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="script_content"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Script Content</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="#!/bin/bash"
                      className="font-mono text-sm"
                      rows={10}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    For full editor experience, save this template and edit from the detail page.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="config_schema"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Config Schema (JSON)</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='{"timeout": {"type": "integer", "default": 30}}'
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
              name="enabled"
              render={({ field }) => (
                <FormItem className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <FormLabel>Enabled</FormLabel>
                    <FormDescription>
                      Enable or disable this script template
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
