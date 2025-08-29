// @ts-nocheck
'use client'

import dynamic from 'next/dynamic'
import React, { useEffect, useState, type FC } from 'react'
import { Loading } from '@/components/Loading'
import CodeEditor from '@uiw/react-textarea-code-editor'
const JsonViewer = dynamic(
  () => import('@andypf/json-viewer/dist/esm/react/JsonViewer'),
  {
    loading: () => (
      <div className="border rounded-md overflow-auto p-1.5">
        <Loading loadingText="Loading JSON viewer..." />
      </div>
    ),
    ssr: false,
  }
) as typeof import('@andypf/json-viewer/dist/esm/react/JsonViewer')

export interface ICodeViewer {
  isEditable?: boolean
  initCodeSource?: string
  language?: 'shell' | 'toml' | 'json' | 'hcl' | 'yaml'
  placeholder?: string
  name?: string
  required?: boolean
}

export const CodeViewer: FC<ICodeViewer> = ({
  isEditable = false,
  initCodeSource = '',
  language = 'shell',
  placeholder = '',
  name,
  required,
}) => {
  const [code, setCode] = useState(initCodeSource)

  return (
    <div className="rounded overflow-auto">
      <CodeEditor
        autoCapitalize="off"
        value={code}
        language={language}
        placeholder={placeholder}
        onChange={(evn) => {
          if (isEditable) setCode(evn.target.value)
        }}
        padding={16}
        readOnly={!isEditable}
        name={name}
        required={required}
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

export const JsonView = ({ data, ...props }) => {
  const [theme, setTheme] = useState('google-dark')

  useEffect(() => {
    if (typeof window !== 'undefined') {
      setTheme(
        window &&
          window?.matchMedia &&
          window.matchMedia('(prefers-color-scheme: dark)').matches
          ? 'google-dark'
          : 'google-light'
      )
    }
  }, [])

  return (
    <>
      <div className="border rounded-md overflow-auto">
        <JsonViewer data={data} {...props} theme={theme} />
      </div>
    </>
  )
}
