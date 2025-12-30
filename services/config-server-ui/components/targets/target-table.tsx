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
import { MoreHorizontal, Edit, Trash2, Eye } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import type { Target } from '@/types/api'

interface TargetTableProps {
  targets: Target[]
  isLoading: boolean
  onEdit: (target: Target) => void
  onDelete: (target: Target) => void
}

export function TargetTable({ targets, isLoading, onEdit, onDelete }: TargetTableProps) {
  if (isLoading) {
    return <TableLoading />
  }

  if (!targets.length) {
    return (
      <div className="flex h-32 items-center justify-center rounded-lg border text-gray-500">
        No targets found. Create your first target to get started.
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Hostname</TableHead>
            <TableHead>IP Address</TableHead>
            <TableHead>Port</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Groups</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {targets.map((target) => (
            <TableRow key={target.id}>
              <TableCell>
                <Link
                  href={`/targets/${target.id}`}
                  className="font-medium hover:underline"
                >
                  {target.hostname}
                </Link>
              </TableCell>
              <TableCell className="font-mono text-sm">{target.ip_address}</TableCell>
              <TableCell>{target.port}</TableCell>
              <TableCell>
                <StatusBadge status={target.status} />
              </TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  {target.groups?.slice(0, 3).map((group) => (
                    <Badge key={group.id} variant="outline">
                      {group.name}
                    </Badge>
                  ))}
                  {target.groups && target.groups.length > 3 && (
                    <Badge variant="secondary">+{target.groups.length - 3}</Badge>
                  )}
                </div>
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem asChild>
                      <Link href={`/targets/${target.id}`}>
                        <Eye className="mr-2 h-4 w-4" />
                        View
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => onEdit(target)}>
                      <Edit className="mr-2 h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => onDelete(target)}
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

function StatusBadge({ status }: { status: Target['status'] }) {
  const variants: Record<Target['status'], 'default' | 'secondary' | 'destructive' | 'outline'> = {
    active: 'default',
    inactive: 'secondary',
    down: 'destructive',
  }

  return <Badge variant={variants[status] || 'outline'}>{status}</Badge>
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
