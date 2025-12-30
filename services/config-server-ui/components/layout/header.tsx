'use client'

import { AlertBadge } from '@/components/alerts/alert-badge'
import { Breadcrumb } from '@/components/layout/breadcrumb'

export function Header() {
  return (
    <header className="flex h-16 items-center justify-between border-b bg-white px-6">
      <Breadcrumb />
      <div className="flex items-center gap-4">
        <AlertBadge />
      </div>
    </header>
  )
}
