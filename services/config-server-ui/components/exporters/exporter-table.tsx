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
import { MoreHorizontal, Edit, Trash2, Power, PowerOff } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import type { Exporter } from '@/types/api'

interface ExporterTableProps {
  exporters: Exporter[]
  isLoading: boolean
  onEdit: (exporter: Exporter) => void
  onDelete: (exporter: Exporter) => void
  onToggle: (exporter: Exporter) => void
}

export function ExporterTable({ exporters, isLoading, onEdit, onDelete, onToggle }: ExporterTableProps) {
  if (isLoading) {
    return <TableLoading />
  }

  if (!exporters.length) {
    return (
      <div className="flex h-32 items-center justify-center rounded-lg border text-gray-500">
        No exporters found.
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Type</TableHead>
            <TableHead>Target</TableHead>
            <TableHead>Port</TableHead>
            <TableHead>Path</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {exporters.map((exporter) => (
            <TableRow key={exporter.id}>
              <TableCell>
                <Badge variant="outline">{exporter.type}</Badge>
              </TableCell>
              <TableCell>
                {exporter.target ? (
                  <Link
                    href={`/targets/${exporter.target_id}`}
                    className="font-medium hover:underline"
                  >
                    {exporter.target.hostname}
                  </Link>
                ) : (
                  <span className="text-gray-500">{exporter.target_id}</span>
                )}
              </TableCell>
              <TableCell>{exporter.port}</TableCell>
              <TableCell className="font-mono text-sm">{exporter.path || '/metrics'}</TableCell>
              <TableCell>
                <Badge variant={exporter.enabled ? 'default' : 'secondary'}>
                  {exporter.enabled ? 'Enabled' : 'Disabled'}
                </Badge>
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => onToggle(exporter)}>
                      {exporter.enabled ? (
                        <>
                          <PowerOff className="mr-2 h-4 w-4" />
                          Disable
                        </>
                      ) : (
                        <>
                          <Power className="mr-2 h-4 w-4" />
                          Enable
                        </>
                      )}
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => onEdit(exporter)}>
                      <Edit className="mr-2 h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => onDelete(exporter)}
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
