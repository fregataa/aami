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
import { MoreHorizontal, Edit, Trash2, FileCode } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import type { ScriptTemplate } from '@/types/api'

interface ScriptTemplateTableProps {
  templates: ScriptTemplate[]
  isLoading: boolean
  onEdit: (template: ScriptTemplate) => void
  onDelete: (template: ScriptTemplate) => void
  onToggleEnabled?: (template: ScriptTemplate, enabled: boolean) => void
}

export function ScriptTemplateTable({
  templates,
  isLoading,
  onEdit,
  onDelete,
  onToggleEnabled,
}: ScriptTemplateTableProps) {
  if (isLoading) {
    return <TableLoading />
  }

  if (!templates.length) {
    return (
      <div className="flex h-32 items-center justify-center rounded-lg border text-gray-500">
        No script templates found. Create your first template to get started.
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Description</TableHead>
            <TableHead>Enabled</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {templates.map((template) => (
            <TableRow key={template.id}>
              <TableCell>
                <Link
                  href={`/scripts/templates/${template.id}`}
                  className="flex items-center gap-2 font-medium hover:underline"
                >
                  <FileCode className="h-4 w-4 text-gray-400" />
                  {template.name}
                </Link>
              </TableCell>
              <TableCell>
                <Badge variant="outline">{template.script_type}</Badge>
              </TableCell>
              <TableCell className="max-w-xs truncate text-gray-500">
                {template.description || '-'}
              </TableCell>
              <TableCell>
                {onToggleEnabled ? (
                  <Switch
                    checked={template.enabled}
                    onCheckedChange={(checked) => onToggleEnabled(template, checked)}
                  />
                ) : (
                  <Badge variant={template.enabled ? 'default' : 'secondary'}>
                    {template.enabled ? 'Yes' : 'No'}
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
                    <DropdownMenuItem onClick={() => onEdit(template)}>
                      <Edit className="mr-2 h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
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

function TableLoading() {
  return (
    <div className="space-y-3">
      {[1, 2, 3, 4, 5].map((i) => (
        <Skeleton key={i} className="h-12 w-full" />
      ))}
    </div>
  )
}
