'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { useSidebar } from './sidebar-provider'
import {
  LayoutDashboard,
  Server,
  FolderTree,
  Gauge,
  Bell,
  FileCode,
  Key,
  Settings,
  X,
} from 'lucide-react'
import { Button } from '@/components/ui/button'

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Targets', href: '/targets', icon: Server },
  { name: 'Groups', href: '/groups', icon: FolderTree },
  { name: 'Exporters', href: '/exporters', icon: Gauge },
  {
    name: 'Alerts',
    children: [
      { name: 'Templates', href: '/alerts/templates' },
      { name: 'Rules', href: '/alerts/rules' },
    ],
    icon: Bell,
  },
  {
    name: 'Scripts',
    children: [
      { name: 'Templates', href: '/scripts/templates' },
      { name: 'Policies', href: '/scripts/policies' },
    ],
    icon: FileCode,
  },
  { name: 'Bootstrap', href: '/bootstrap', icon: Key },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export function Sidebar() {
  const pathname = usePathname()
  const { isOpen, close } = useSidebar()

  return (
    <>
      {/* Mobile overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={close}
        />
      )}

      {/* Sidebar */}
      <div
        className={cn(
          'fixed inset-y-0 left-0 z-50 w-64 transform border-r bg-white transition-transform duration-200 ease-in-out lg:relative lg:translate-x-0',
          isOpen ? 'translate-x-0' : '-translate-x-full'
        )}
      >
        <div className="flex h-16 items-center justify-between border-b px-6">
          <span className="text-xl font-bold">AAMI</span>
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            onClick={close}
          >
            <X className="h-5 w-5" />
          </Button>
        </div>
        <nav className="flex-1 space-y-1 overflow-y-auto p-4">
          {navigation.map((item) =>
            item.children ? (
              <div key={item.name} className="space-y-1">
                <div className="flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-500">
                  <item.icon className="h-4 w-4" />
                  {item.name}
                </div>
                {item.children.map((child) => (
                  <Link
                    key={child.href}
                    href={child.href}
                    onClick={close}
                    className={cn(
                      'block rounded-md px-3 py-2 pl-9 text-sm',
                      pathname === child.href || pathname.startsWith(child.href + '/')
                        ? 'bg-gray-100 font-medium text-gray-900'
                        : 'text-gray-600 hover:bg-gray-50'
                    )}
                  >
                    {child.name}
                  </Link>
                ))}
              </div>
            ) : (
              <Link
                key={item.href}
                href={item.href}
                onClick={close}
                className={cn(
                  'flex items-center gap-2 rounded-md px-3 py-2 text-sm',
                  pathname === item.href || (item.href !== '/' && pathname.startsWith(item.href))
                    ? 'bg-gray-100 font-medium text-gray-900'
                    : 'text-gray-600 hover:bg-gray-50'
                )}
              >
                <item.icon className="h-4 w-4" />
                {item.name}
              </Link>
            )
          )}
        </nav>
      </div>
    </>
  )
}
