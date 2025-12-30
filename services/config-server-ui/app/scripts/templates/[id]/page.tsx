'use client'

import { use, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useScriptTemplate } from '@/lib/hooks/use-script-templates'
import { useScriptPoliciesByTemplate } from '@/lib/hooks/use-script-policies'
import { scriptTemplatesApi, type UpdateScriptTemplateRequest } from '@/lib/api/script-templates'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { ScriptEditor } from '@/components/scripts/script-editor'
import { DeleteDialog } from '@/components/shared/delete-dialog'
import { toast } from 'sonner'
import { ArrowLeft, Save, RotateCcw, Trash2, CheckCircle, XCircle } from 'lucide-react'

export default function ScriptTemplateDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  const router = useRouter()
  const { template, isLoading, mutate } = useScriptTemplate(id)
  const { policies } = useScriptPoliciesByTemplate(id)

  const [editedContent, setEditedContent] = useState<string | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [hashValid, setHashValid] = useState<boolean | null>(null)

  const hasChanges = editedContent !== null && editedContent !== template?.script_content

  const handleSave = async () => {
    if (!hasChanges || editedContent === null) return

    setIsSaving(true)
    try {
      await scriptTemplatesApi.update(id, { script_content: editedContent })
      toast.success('Script saved successfully')
      mutate()
      setEditedContent(null)
      setHashValid(null)
    } catch (error) {
      toast.error('Failed to save script')
    } finally {
      setIsSaving(false)
    }
  }

  const handleReset = () => {
    setEditedContent(null)
  }

  const handleToggleEnabled = async (enabled: boolean) => {
    try {
      await scriptTemplatesApi.update(id, { enabled })
      toast.success(`Script template ${enabled ? 'enabled' : 'disabled'}`)
      mutate()
    } catch (error) {
      toast.error('Failed to update script template')
    }
  }

  const handleVerifyHash = async () => {
    try {
      const result = await scriptTemplatesApi.verifyHash(id)
      setHashValid(result.valid)
      toast[result.valid ? 'success' : 'error'](
        result.valid ? 'Hash is valid' : 'Hash mismatch detected'
      )
    } catch (error) {
      toast.error('Failed to verify hash')
    }
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      await scriptTemplatesApi.delete(id)
      toast.success('Script template deleted successfully')
      router.push('/scripts/templates')
    } catch (error) {
      toast.error('Failed to delete script template')
      setIsDeleting(false)
    }
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-96 w-full" />
      </div>
    )
  }

  if (!template) {
    return (
      <div className="space-y-6">
        <Link href="/scripts/templates">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Templates
          </Button>
        </Link>
        <div className="text-center text-gray-500">Template not found</div>
      </div>
    )
  }

  const currentContent = editedContent ?? template.script_content

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link href="/scripts/templates">
            <Button variant="ghost" size="icon">
              <ArrowLeft className="h-4 w-4" />
            </Button>
          </Link>
          <h1 className="text-2xl font-bold">{template.name}</h1>
          <Badge variant="outline">{template.script_type}</Badge>
          <Badge variant={template.enabled ? 'default' : 'secondary'}>
            {template.enabled ? 'Enabled' : 'Disabled'}
          </Badge>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleVerifyHash}>
            {hashValid === true && <CheckCircle className="mr-2 h-4 w-4 text-green-600" />}
            {hashValid === false && <XCircle className="mr-2 h-4 w-4 text-red-600" />}
            Verify Hash
          </Button>
          <Button variant="outline" onClick={() => setShowDelete(true)}>
            <Trash2 className="mr-2 h-4 w-4" />
            Delete
          </Button>
          {hasChanges && (
            <>
              <Button variant="outline" onClick={handleReset}>
                <RotateCcw className="mr-2 h-4 w-4" />
                Reset
              </Button>
              <Button onClick={handleSave} disabled={isSaving}>
                <Save className="mr-2 h-4 w-4" />
                {isSaving ? 'Saving...' : 'Save'}
              </Button>
            </>
          )}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Script Content</CardTitle>
        </CardHeader>
        <CardContent>
          <ScriptEditor
            value={currentContent}
            onChange={setEditedContent}
            language={template.script_type === 'python' ? 'python' : 'shell'}
            height="500px"
          />
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Details</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="text-sm font-medium text-gray-500">Enabled</div>
              <Switch checked={template.enabled} onCheckedChange={handleToggleEnabled} />
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Description</div>
              <div>{template.description || '-'}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Hash</div>
              <code className="break-all text-sm">{template.hash}</code>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Created</div>
              <div>{new Date(template.created_at).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-500">Updated</div>
              <div>{new Date(template.updated_at).toLocaleString()}</div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Config Schema</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="overflow-x-auto rounded bg-gray-100 p-3 text-sm">
              {JSON.stringify(template.config_schema, null, 2)}
            </pre>
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Policies Using This Template ({policies.length})</CardTitle>
          </CardHeader>
          <CardContent>
            {policies.length > 0 ? (
              <div className="space-y-2">
                {policies.map((policy) => (
                  <Link
                    key={policy.id}
                    href={`/scripts/policies/${policy.id}`}
                    className="block rounded-lg border p-3 hover:bg-gray-50"
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">
                          {policy.group?.name ?? 'Global Policy'}
                        </div>
                        <div className="text-sm text-gray-500">
                          Priority: {policy.priority}
                        </div>
                      </div>
                      <Badge variant={policy.enabled ? 'default' : 'secondary'}>
                        {policy.enabled ? 'Enabled' : 'Disabled'}
                      </Badge>
                    </div>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="text-gray-500">No policies using this template</div>
            )}
          </CardContent>
        </Card>
      </div>

      <DeleteDialog
        open={showDelete}
        onClose={() => setShowDelete(false)}
        onConfirm={handleDelete}
        title="Delete Script Template"
        description={`Are you sure you want to delete "${template.name}"? This action cannot be undone.`}
        isDeleting={isDeleting}
      />
    </div>
  )
}
