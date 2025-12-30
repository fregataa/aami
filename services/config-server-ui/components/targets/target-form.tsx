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
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { GroupMultiSelect } from '@/components/groups/group-multi-select'
import type { Target } from '@/types/api'

const formSchema = z.object({
  hostname: z.string().min(1, 'Hostname is required'),
  ip_address: z.string().min(1, 'IP address is required'),
  port: z.number().min(1).max(65535),
  group_ids: z.array(z.string()),
})

type FormValues = z.infer<typeof formSchema>

interface TargetFormProps {
  open: boolean
  onClose: () => void
  target?: Target | null
  onSubmit: (data: FormValues) => Promise<void>
}

export function TargetForm({ open, onClose, target, onSubmit }: TargetFormProps) {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      hostname: '',
      ip_address: '',
      port: 9100,
      group_ids: [],
    },
  })

  useEffect(() => {
    if (target) {
      form.reset({
        hostname: target.hostname,
        ip_address: target.ip_address,
        port: target.port,
        group_ids: target.groups?.map((g) => g.id) ?? [],
      })
    } else {
      form.reset({
        hostname: '',
        ip_address: '',
        port: 9100,
        group_ids: [],
      })
    }
  }, [target, form])

  const handleSubmit = async (data: FormValues) => {
    await onSubmit(data)
    form.reset()
  }

  const handleClose = () => {
    form.reset()
    onClose()
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{target ? 'Edit Target' : 'Create Target'}</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="hostname"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Hostname</FormLabel>
                  <FormControl>
                    <Input placeholder="server-01.example.com" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="ip_address"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>IP Address</FormLabel>
                  <FormControl>
                    <Input placeholder="192.168.1.100" {...field} />
                  </FormControl>
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
                      placeholder="9100"
                      {...field}
                      onChange={(e) => field.onChange(parseInt(e.target.value) || 9100)}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="group_ids"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Groups</FormLabel>
                  <FormControl>
                    <GroupMultiSelect
                      value={field.value}
                      onChange={field.onChange}
                    />
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
