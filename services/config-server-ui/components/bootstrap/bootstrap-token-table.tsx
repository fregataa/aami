'use client'

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
import { MoreHorizontal, Edit, Trash2, Copy, Check } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import { useState } from 'react'
import { toast } from 'sonner'
import { formatDistanceToNow, isPast, parseISO } from 'date-fns'
import type { BootstrapToken } from '@/types/api'

interface BootstrapTokenTableProps {
  tokens: BootstrapToken[]
  isLoading: boolean
  onEdit: (token: BootstrapToken) => void
  onDelete: (token: BootstrapToken) => void
  onSelect: (token: BootstrapToken) => void
}

export function BootstrapTokenTable({
  tokens,
  isLoading,
  onEdit,
  onDelete,
  onSelect,
}: BootstrapTokenTableProps) {
  const [copiedId, setCopiedId] = useState<string | null>(null)

  const handleCopyToken = async (token: BootstrapToken) => {
    if (!token.token) {
      toast.error('Token value is not available')
      return
    }
    try {
      await navigator.clipboard.writeText(token.token)
      setCopiedId(token.id)
      toast.success('Token copied to clipboard')
      setTimeout(() => setCopiedId(null), 2000)
    } catch {
      toast.error('Failed to copy token')
    }
  }

  if (isLoading) {
    return <TableLoading />
  }

  if (!tokens.length) {
    return (
      <div className="flex h-32 items-center justify-center rounded-lg border text-gray-500">
        No bootstrap tokens found. Create your first token to get started.
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Group</TableHead>
            <TableHead>Expires</TableHead>
            <TableHead>Usage</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {tokens.map((token) => {
            const isExpired = token.expires_at ? isPast(parseISO(token.expires_at)) : false
            const isExhausted = token.max_uses > 0 && token.use_count >= token.max_uses

            return (
              <TableRow
                key={token.id}
                className="cursor-pointer hover:bg-muted/50"
                onClick={() => onSelect(token)}
              >
                <TableCell>
                  <div>
                    <div className="font-medium">{token.name}</div>
                    {token.description && (
                      <div className="text-sm text-gray-500 truncate max-w-[200px]">
                        {token.description}
                      </div>
                    )}
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant="outline">{token.group?.name || 'N/A'}</Badge>
                </TableCell>
                <TableCell>
                  {token.expires_at ? (
                    <span className={isExpired ? 'text-red-600' : ''}>
                      {isExpired
                        ? 'Expired'
                        : formatDistanceToNow(parseISO(token.expires_at), { addSuffix: true })}
                    </span>
                  ) : (
                    <span className="text-gray-500">Never</span>
                  )}
                </TableCell>
                <TableCell>
                  <span className={isExhausted ? 'text-red-600' : ''}>
                    {token.use_count} / {token.max_uses > 0 ? token.max_uses : 'unlimited'}
                  </span>
                </TableCell>
                <TableCell>
                  <TokenStatusBadge isExpired={isExpired} isExhausted={isExhausted} />
                </TableCell>
                <TableCell>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                      <Button variant="ghost" size="icon">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      {token.token && (
                        <DropdownMenuItem
                          onClick={(e) => {
                            e.stopPropagation()
                            handleCopyToken(token)
                          }}
                        >
                          {copiedId === token.id ? (
                            <Check className="mr-2 h-4 w-4" />
                          ) : (
                            <Copy className="mr-2 h-4 w-4" />
                          )}
                          Copy Token
                        </DropdownMenuItem>
                      )}
                      <DropdownMenuItem
                        onClick={(e) => {
                          e.stopPropagation()
                          onEdit(token)
                        }}
                      >
                        <Edit className="mr-2 h-4 w-4" />
                        Edit
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={(e) => {
                          e.stopPropagation()
                          onDelete(token)
                        }}
                        className="text-red-600"
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}

function TokenStatusBadge({
  isExpired,
  isExhausted,
}: {
  isExpired?: boolean
  isExhausted?: boolean
}) {
  if (isExpired) {
    return <Badge variant="destructive">Expired</Badge>
  }
  if (isExhausted) {
    return <Badge variant="secondary">Exhausted</Badge>
  }
  return <Badge variant="default">Active</Badge>
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
