'use client'

import { ExternalLink } from 'lucide-react'

const links = [
  { name: 'Grafana', url: 'http://localhost:3000', description: 'Dashboards & Visualization' },
  { name: 'Prometheus', url: 'http://localhost:9090', description: 'Metrics & Queries' },
  { name: 'Alertmanager', url: 'http://localhost:9093', description: 'Alert Management' },
]

export function ExternalLinks() {
  return (
    <div className="space-y-3">
      {links.map((link) => (
        <a
          key={link.name}
          href={link.url}
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center justify-between rounded-lg border p-4 hover:bg-gray-50"
        >
          <div>
            <div className="font-medium">{link.name}</div>
            <div className="text-sm text-gray-500">{link.description}</div>
          </div>
          <ExternalLink className="h-4 w-4 text-gray-400" />
        </a>
      ))}
    </div>
  )
}
