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
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useTargets } from '@/lib/hooks/use-targets'
import type { Exporter } from '@/types/api'

const EXPORTER_TYPES = [
  { value: 'node_exporter', label: 'Node Exporter' },
  { value: 'dcgm_exporter', label: 'DCGM Exporter (GPU)' },
  { value: 'lustre_exporter', label: 'Lustre Exporter' },
  { value: 'infiniband_exporter', label: 'InfiniBand Exporter' },
  { value: 'nvmeof_exporter', label: 'NVMe-oF Exporter' },
  { value: 'custom', label: 'Custom' },
]

const formSchema = z.object({
  target_id: z.string().min(1, 'Target is required'),
  type: z.string().min(1, 'Type is required'),
  port: z.number().min(1).max(65535),
  path: z.string(),
})

type FormValues = z.infer<typeof formSchema>

interface ExporterFormProps {
  open: boolean
  onClose: () => void
  exporter?: Exporter | null
  onSubmit: (data: FormValues) => Promise<void>
}

export function ExporterForm({ open, onClose, exporter, onSubmit }: ExporterFormProps) {
  const { targets } = useTargets()

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      target_id: '',
      type: 'node_exporter',
      port: 9100,
      path: '/metrics',
    },
  })

  useEffect(() => {
    if (exporter) {
      form.reset({
        target_id: exporter.target_id,
        type: exporter.type,
        port: exporter.port,
        path: exporter.path || '/metrics',
      })
    } else {
      form.reset({
        target_id: '',
        type: 'node_exporter',
        port: 9100,
        path: '/metrics',
      })
    }
  }, [exporter, form])

  const handleSubmit = async (data: FormValues) => {
    await onSubmit(data)
    form.reset()
  }

  const handleClose = () => {
    form.reset()
    onClose()
  }

  // Update port based on exporter type
  const handleTypeChange = (type: string) => {
    form.setValue('type', type)
    const portMap: Record<string, number> = {
      node_exporter: 9100,
      dcgm_exporter: 9400,
      lustre_exporter: 9169,
      infiniband_exporter: 9315,
      nvmeof_exporter: 9500,
    }
    if (portMap[type]) {
      form.setValue('port', portMap[type])
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{exporter ? 'Edit Exporter' : 'Add Exporter'}</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="target_id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Target</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    value={field.value}
                    disabled={!!exporter}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select target" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {targets.map((target) => (
                        <SelectItem key={target.id} value={target.id}>
                          {target.hostname}
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
              name="type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Exporter Type</FormLabel>
                  <Select
                    onValueChange={handleTypeChange}
                    value={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select type" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {EXPORTER_TYPES.map((type) => (
                        <SelectItem key={type.value} value={type.value}>
                          {type.label}
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
              name="port"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Port</FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      {...field}
                      onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="path"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Metrics Path</FormLabel>
                  <FormControl>
                    <Input placeholder="/metrics" {...field} />
                  </FormControl>
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
