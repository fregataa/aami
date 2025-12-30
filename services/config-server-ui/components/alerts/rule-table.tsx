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
import type { AlertRule } from '@/types/api'

interface AlertRuleTableProps {
  rules: AlertRule[]
  isLoading: boolean
  onEdit: (rule: AlertRule) => void
  onDelete: (rule: AlertRule) => void
  onToggleEnabled?: (rule: AlertRule, enabled: boolean) => void
}

export function AlertRuleTable({
  rules,
  isLoading,
  onEdit,
  onDelete,
  onToggleEnabled,
}: AlertRuleTableProps) {
  if (isLoading) {
    return <TableLoading />
  }

  if (!rules.length) {
    return (
      <div className="flex h-32 items-center justify-center rounded-lg border text-gray-500">
        No alert rules found. Create your first rule to get started.
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Severity</TableHead>
            <TableHead>Group</TableHead>
            <TableHead>Template</TableHead>
            <TableHead>Enabled</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rules.map((rule) => (
            <TableRow key={rule.id}>
              <TableCell>
                <Link
                  href={`/alerts/rules/${rule.id}`}
                  className="font-medium hover:underline"
                >
                  {rule.name}
                </Link>
              </TableCell>
              <TableCell>
                <SeverityBadge severity={rule.severity} />
              </TableCell>
              <TableCell>
                {rule.group ? (
                  <Link
                    href={`/groups/${rule.group.id}`}
                    className="text-blue-600 hover:underline"
                  >
                    {rule.group.name}
                  </Link>
                ) : (
                  <span className="text-gray-400">-</span>
                )}
              </TableCell>
              <TableCell>
                {rule.created_from_template_name ? (
                  <Link
                    href={`/alerts/templates/${rule.created_from_template_id}`}
                    className="text-blue-600 hover:underline"
                  >
                    {rule.created_from_template_name}
                  </Link>
                ) : (
                  <span className="text-gray-400">Custom</span>
                )}
              </TableCell>
              <TableCell>
                {onToggleEnabled ? (
                  <Switch
                    checked={rule.enabled}
                    onCheckedChange={(checked) => onToggleEnabled(rule, checked)}
                  />
                ) : (
                  <Badge variant={rule.enabled ? 'default' : 'secondary'}>
                    {rule.enabled ? 'Yes' : 'No'}
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
                    <DropdownMenuItem onClick={() => onEdit(rule)}>
                      <Edit className="mr-2 h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => onDelete(rule)}
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

function SeverityBadge({ severity }: { severity: AlertRule['severity'] }) {
  const variants: Record<AlertRule['severity'], 'default' | 'secondary' | 'destructive'> = {
    critical: 'destructive',
    warning: 'default',
    info: 'secondary',
  }

  return <Badge variant={variants[severity]}>{severity}</Badge>
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
