'use client'

import { useState, useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Copy, Check, Terminal } from 'lucide-react'
import { toast } from 'sonner'
import type { BootstrapToken } from '@/types/api'

interface CommandGeneratorProps {
  token: BootstrapToken | null
  serverUrl?: string
}

export function CommandGenerator({ token, serverUrl: initialServerUrl }: CommandGeneratorProps) {
  const [serverUrl, setServerUrl] = useState(initialServerUrl || 'https://config.example.com')
  const [labels, setLabels] = useState('')
  const [copied, setCopied] = useState(false)

  const command = useMemo(() => {
    if (!token?.token) return ''

    let cmd = `curl -fsSL ${serverUrl}/api/v1/bootstrap/script | sudo bash -s -- \\
  --token ${token.token} \\
  --server ${serverUrl}`

    if (labels.trim()) {
      const labelParts = labels.split(',').map((l) => l.trim()).filter(Boolean)
      labelParts.forEach((label) => {
        cmd += ` \\
  --labels ${label}`
      })
    }

    return cmd
  }, [token, serverUrl, labels])

  const handleCopy = async () => {
    if (!command) return

    try {
      await navigator.clipboard.writeText(command)
      setCopied(true)
      toast.success('Command copied to clipboard')
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error('Failed to copy')
    }
  }

  if (!token) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Terminal className="h-5 w-5" />
            Bootstrap Command
          </CardTitle>
          <CardDescription>
            Select a token from the table to generate the bootstrap command.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex h-32 items-center justify-center rounded-lg border border-dashed text-gray-500">
            Select a bootstrap token to generate the command
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Terminal className="h-5 w-5" />
          Bootstrap Command
        </CardTitle>
        <CardDescription>
          Run this command on target nodes to register them with the config server.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="server-url">Config Server URL</Label>
            <Input
              id="server-url"
              value={serverUrl}
              onChange={(e) => setServerUrl(e.target.value)}
              placeholder="https://config.example.com"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="labels">Labels (comma-separated)</Label>
            <Input
              id="labels"
              value={labels}
              onChange={(e) => setLabels(e.target.value)}
              placeholder="env=production, datacenter=dc1"
            />
          </div>
        </div>

        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <Label>Generated Command</Label>
            <Button
              variant="outline"
              size="sm"
              onClick={handleCopy}
              disabled={!command}
            >
              {copied ? (
                <>
                  <Check className="mr-2 h-4 w-4" />
                  Copied
                </>
              ) : (
                <>
                  <Copy className="mr-2 h-4 w-4" />
                  Copy
                </>
              )}
            </Button>
          </div>
          <div className="relative">
            <pre className="overflow-x-auto rounded-lg bg-gray-900 p-4 text-sm text-gray-100">
              {command || 'Token value not available'}
            </pre>
          </div>
        </div>

        <div className="rounded-lg bg-blue-50 p-4 text-sm text-blue-800">
          <strong>Token:</strong> {token.name}
          <br />
          <strong>Group:</strong> {token.group?.name || 'N/A'}
          <br />
          <strong>Remaining uses:</strong>{' '}
          {token.max_uses > 0
            ? `${token.max_uses - token.use_count} of ${token.max_uses}`
            : 'Unlimited'}
        </div>
      </CardContent>
    </Card>
  )
}
