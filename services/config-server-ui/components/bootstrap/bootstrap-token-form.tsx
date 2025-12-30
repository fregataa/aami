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
  DialogDescription,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { GroupSelect } from '@/components/groups/group-select'
import type { BootstrapToken } from '@/types/api'

const formSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
  group_id: z.string().min(1, 'Group is required'),
  expires_in_days: z.number().min(0).optional(),
  max_uses: z.number().min(0).optional(),
})

type FormValues = z.infer<typeof formSchema>

interface BootstrapTokenFormProps {
  open: boolean
  onClose: () => void
  token?: BootstrapToken | null
  onSubmit: (data: FormValues) => Promise<void>
}

export function BootstrapTokenForm({ open, onClose, token, onSubmit }: BootstrapTokenFormProps) {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      description: '',
      group_id: '',
      expires_in_days: 30,
      max_uses: 0,
    },
  })

  useEffect(() => {
    if (token) {
      form.reset({
        name: token.name,
        description: token.description || '',
        group_id: token.group_id,
        expires_in_days: undefined,
        max_uses: token.max_uses,
      })
    } else {
      form.reset({
        name: '',
        description: '',
        group_id: '',
        expires_in_days: 30,
        max_uses: 0,
      })
    }
  }, [token, form])

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
          <DialogTitle>{token ? 'Edit Bootstrap Token' : 'Create Bootstrap Token'}</DialogTitle>
          <DialogDescription>
            {token
              ? 'Update the bootstrap token settings.'
              : 'Create a new bootstrap token for node registration.'}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder="production-token" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Input placeholder="Token for production servers" {...field} />
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
                  <FormLabel>Target Group</FormLabel>
                  <FormControl>
                    <GroupSelect
                      value={field.value}
                      onChange={field.onChange}
                      placeholder="Select target group"
                    />
                  </FormControl>
                  <FormDescription>
                    Nodes registered with this token will be added to this group.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            {!token && (
              <FormField
                control={form.control}
                name="expires_in_days"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Expires In (days)</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        placeholder="30"
                        {...field}
                        onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                      />
                    </FormControl>
                    <FormDescription>
                      Set to 0 for no expiration.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            <FormField
              control={form.control}
              name="max_uses"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Max Uses</FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      placeholder="0"
                      {...field}
                      onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                    />
                  </FormControl>
                  <FormDescription>
                    Maximum number of times this token can be used. Set to 0 for unlimited.
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
