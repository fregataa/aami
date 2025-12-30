'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { ChevronRight, Home } from 'lucide-react'

export function Breadcrumb() {
  const pathname = usePathname()
  const segments = pathname.split('/').filter(Boolean)

  const breadcrumbs = segments.map((segment, index) => {
    const path = '/' + segments.slice(0, index + 1).join('/')
    const label = segment.charAt(0).toUpperCase() + segment.slice(1).replace(/-/g, ' ')
    return { path, label }
  })

  return (
    <nav className="flex items-center gap-2 text-sm text-gray-500">
      <Link href="/" className="hover:text-gray-900">
        <Home className="h-4 w-4" />
      </Link>
      {breadcrumbs.map((crumb, index) => (
        <div key={crumb.path} className="flex items-center gap-2">
          <ChevronRight className="h-4 w-4" />
          {index === breadcrumbs.length - 1 ? (
            <span className="font-medium text-gray-900">{crumb.label}</span>
          ) : (
            <Link href={crumb.path} className="hover:text-gray-900">
              {crumb.label}
            </Link>
          )}
        </div>
      ))}
    </nav>
  )
}
