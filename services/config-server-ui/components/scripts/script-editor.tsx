'use client'

import { useRef, useCallback } from 'react'
import Editor, { OnMount } from '@monaco-editor/react'
import type { editor } from 'monaco-editor'

interface ScriptEditorProps {
  value: string
  onChange: (value: string) => void
  language?: string
  readOnly?: boolean
  height?: string
}

export function ScriptEditor({
  value,
  onChange,
  language = 'shell',
  readOnly = false,
  height = '400px',
}: ScriptEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null)

  const handleMount: OnMount = useCallback((editor) => {
    editorRef.current = editor
  }, [])

  const handleChange = useCallback(
    (value: string | undefined) => {
      onChange(value ?? '')
    },
    [onChange]
  )

  return (
    <div className="overflow-hidden rounded-lg border">
      <Editor
        height={height}
        language={language}
        value={value}
        onChange={handleChange}
        onMount={handleMount}
        options={{
          readOnly,
          minimap: { enabled: false },
          lineNumbers: 'on',
          scrollBeyondLastLine: false,
          fontSize: 14,
          tabSize: 2,
          automaticLayout: true,
          wordWrap: 'on',
        }}
        theme="vs-light"
      />
    </div>
  )
}
