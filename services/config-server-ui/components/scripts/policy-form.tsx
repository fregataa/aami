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
import { useScriptTemplates } from '@/lib/hooks/use-script-templates'
import type { ScriptPolicy } from '@/types/api'

const formSchema = z.object({
  template_id: z.string().min(1, 'Template is required'),
  group_id: z.string(),
  config: z.string(),
  priority: z.number().min(0),
  enabled: z.boolean(),
})

type FormValues = z.infer<typeof formSchema>

interface ScriptPolicyFormProps {
  open: boolean
  onClose: () => void
  policy?: ScriptPolicy | null
  onSubmit: (data: {
    template_id: string
    group_id?: string
    config?: Record<string, unknown>
    priority: number
    enabled: boolean
  }) => Promise<void>
}

export function ScriptPolicyForm({ open, onClose, policy, onSubmit }: ScriptPolicyFormProps) {
  const { groups } = useGroups()
  const { templates } = useScriptTemplates()

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      template_id: '',
      group_id: '',
      config: '{}',
      priority: 0,
      enabled: true,
    },
  })

  useEffect(() => {
    if (policy) {
      form.reset({
        template_id: policy.template_id,
        group_id: policy.group_id || '',
        config: JSON.stringify(policy.config || {}, null, 2),
        priority: policy.priority,
        enabled: policy.enabled,
      })
    } else {
      form.reset({
        template_id: '',
        group_id: '',
        config: '{}',
        priority: 0,
        enabled: true,
      })
    }
  }, [policy, form])

  const handleSubmit = async (data: FormValues) => {
    let config: Record<string, unknown> = {}
    try {
      config = JSON.parse(data.config)
    } catch {
      form.setError('config', { message: 'Invalid JSON' })
      return
    }

    await onSubmit({
      template_id: data.template_id,
      group_id: data.group_id || undefined,
      config: config,
      priority: data.priority,
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
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{policy ? 'Edit Script Policy' : 'Create Script Policy'}</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="template_id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Script Template</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select template" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {templates.map((template) => (
                        <SelectItem key={template.id} value={template.id}>
                          {template.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="group_id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Group (Optional)</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Global (all targets)" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="">Global (all targets)</SelectItem>
                      {groups.map((group) => (
                        <SelectItem key={group.id} value={group.id}>
                          {group.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    Leave empty to apply to all targets
                  </FormDescription>
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
                  <FormDescription>
                    Higher priority policies take precedence
                  </FormDescription>
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
                      placeholder='{"timeout": 60}'
                      className="font-mono text-sm"
                      rows={4}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Override template config values for this policy
                  </FormDescription>
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
                      Enable or disable this policy
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
