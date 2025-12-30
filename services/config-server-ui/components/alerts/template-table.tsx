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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { MoreHorizontal, Edit, Trash2, Copy } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import type { AlertTemplate } from '@/types/api'

interface AlertTemplateTableProps {
  templates: AlertTemplate[]
  isLoading: boolean
  onEdit: (template: AlertTemplate) => void
  onDelete: (template: AlertTemplate) => void
  onDuplicate?: (template: AlertTemplate) => void
}

export function AlertTemplateTable({
  templates,
  isLoading,
  onEdit,
  onDelete,
  onDuplicate,
}: AlertTemplateTableProps) {
  if (isLoading) {
    return <TableLoading />
  }

  if (!templates.length) {
    return (
      <div className="flex h-32 items-center justify-center rounded-lg border text-gray-500">
        No alert templates found. Create your first template to get started.
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
            <TableHead>Description</TableHead>
            <TableHead>Created</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {templates.map((template) => (
            <TableRow key={template.id}>
              <TableCell>
                <Link
                  href={`/alerts/templates/${template.id}`}
                  className="font-medium hover:underline"
                >
                  {template.name}
                </Link>
              </TableCell>
              <TableCell>
                <SeverityBadge severity={template.severity} />
              </TableCell>
              <TableCell className="max-w-xs truncate text-gray-500">
                {template.description || '-'}
              </TableCell>
              <TableCell className="text-gray-500">
                {new Date(template.created_at).toLocaleDateString()}
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => onEdit(template)}>
                      <Edit className="mr-2 h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    {onDuplicate && (
                      <DropdownMenuItem onClick={() => onDuplicate(template)}>
                        <Copy className="mr-2 h-4 w-4" />
                        Duplicate
                      </DropdownMenuItem>
                    )}
                    <DropdownMenuItem
                      onClick={() => onDelete(template)}
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

function SeverityBadge({ severity }: { severity: AlertTemplate['severity'] }) {
  const variants: Record<AlertTemplate['severity'], 'default' | 'secondary' | 'destructive'> = {
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
