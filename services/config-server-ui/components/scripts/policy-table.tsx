'use client'

import Link from 'next/link'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { MoreHorizontal, Edit, Trash2 } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import type { ScriptPolicy } from '@/types/api'

interface ScriptPolicyTableProps {
  policies: ScriptPolicy[]
  isLoading: boolean
  onEdit: (policy: ScriptPolicy) => void
  onDelete: (policy: ScriptPolicy) => void
  onToggleEnabled?: (policy: ScriptPolicy, enabled: boolean) => void
}

export function ScriptPolicyTable({
  policies,
  isLoading,
  onEdit,
  onDelete,
  onToggleEnabled,
}: ScriptPolicyTableProps) {
  if (isLoading) {
    return <TableLoading />
  }

  if (!policies.length) {
    return (
      <div className="flex h-32 items-center justify-center rounded-lg border text-gray-500">
        No script policies found. Create your first policy to get started.
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Template</TableHead>
            <TableHead>Group</TableHead>
            <TableHead>Priority</TableHead>
            <TableHead>Enabled</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {policies.map((policy) => (
            <TableRow key={policy.id}>
              <TableCell>
                {policy.template ? (
                  <Link
                    href={`/scripts/templates/${policy.template.id}`}
                    className="font-medium hover:underline"
                  >
                    {policy.template.name}
                  </Link>
                ) : (
                  <span className="text-gray-400">Unknown</span>
                )}
              </TableCell>
              <TableCell>
                {policy.group ? (
                  <Link
                    href={`/groups/${policy.group.id}`}
                    className="text-blue-600 hover:underline"
                  >
                    {policy.group.name}
                  </Link>
                ) : (
                  <Badge variant="secondary">Global</Badge>
                )}
              </TableCell>
              <TableCell>{policy.priority}</TableCell>
              <TableCell>
                {onToggleEnabled ? (
                  <Switch
                    checked={policy.enabled}
                    onCheckedChange={(checked) => onToggleEnabled(policy, checked)}
                  />
                ) : (
                  <Badge variant={policy.enabled ? 'default' : 'secondary'}>
                    {policy.enabled ? 'Yes' : 'No'}
                  </Badge>
                )}
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => onEdit(policy)}>
                      <Edit className="mr-2 h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => onDelete(policy)}
                      className="text-red-600"
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

function TableLoading() {
  return (
    <div className="space-y-3">
      {[1, 2, 3, 4, 5].map((i) => (
        <Skeleton key={i} className="h-12 w-full" />
      ))}
    </div>
  )
}
