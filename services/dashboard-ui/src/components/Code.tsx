'use client'

import React, { useState, type FC } from 'react'
import CodeEditor from '@uiw/react-textarea-code-editor'

export interface ICodeViewer {
  isEditable?: boolean
  initCodeSource?: string
  language?: 'shell' | 'toml' | 'json' | 'hcl' | 'yaml'
  placeholder?: string
}

export const CodeViewer: FC<ICodeViewer> = ({
  isEditable = false,
  initCodeSource = '',
  language = 'shell',
  placeholder = '',
}) => {
  const [code, setCode] = useState(initCodeSource)

  return (
    <div className="rounded overflow-auto">
      <CodeEditor
        value={code}
        language={language}
        placeholder={placeholder}
        onChange={(evn) => {
          if (isEditable) setCode(evn.target.value)
        }}
        padding={16}
        readOnly={!isEditable}
        style={{
          backgroundColor: 'light-dark(#EAEDF0, #19171C)',
          color: 'light-dark(#1E50C0, #6792F4)',
          fontFamily:
            'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
          fontSize: '0.75rem',
          minWidth: '100%',
          width: 'max-content',
          maxWidth: '500px',
        }}
      />
    </div>
  )
}
