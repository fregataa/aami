'use client'

import { AlertBadge } from '@/components/alerts/alert-badge'
import { Breadcrumb } from '@/components/layout/breadcrumb'
import { useSidebar } from './sidebar-provider'
import { Button } from '@/components/ui/button'
import { Menu } from 'lucide-react'

export function Header() {
  const { toggle } = useSidebar()

  return (
    <header className="flex h-16 items-center justify-between border-b bg-white px-4 lg:px-6">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          className="lg:hidden"
          onClick={toggle}
        >
          <Menu className="h-5 w-5" />
        </Button>
        <Breadcrumb />
      </div>
      <div className="flex items-center gap-4">
        <AlertBadge />
      </div>
    </header>
  )
}
