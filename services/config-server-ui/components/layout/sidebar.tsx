'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import {
  LayoutDashboard,
  Server,
  FolderTree,
  Gauge,
  Bell,
  FileCode,
  Key,
  Settings,
} from 'lucide-react'

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

  return (
    <div className="flex h-full w-64 flex-col border-r bg-white">
      <div className="flex h-16 items-center border-b px-6">
        <span className="text-xl font-bold">AAMI</span>
      </div>
      <nav className="flex-1 space-y-1 p-4">
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
  )
}
