'use client'

import Link from 'next/link'
import { Bell } from 'lucide-react'
import { useActiveAlerts } from '@/lib/hooks/use-active-alerts'

export function AlertBadge() {
  const { total, isLoading } = useActiveAlerts()

  return (
    <Link
      href="/alerts/rules"
      className="relative flex items-center gap-2 rounded-md px-3 py-2 text-sm text-gray-600 hover:bg-gray-100"
    >
      <Bell className="h-5 w-5" />
      {!isLoading && total > 0 && (
        <span className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-xs font-medium text-white">
          {total > 99 ? '99+' : total}
        </span>
      )}
    </Link>
  )
}
