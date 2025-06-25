'use client'

import dynamic from 'next/dynamic'
import React from 'react'
import { cn } from '@/stratus/components/helpers'

const MonacoEditor = dynamic(() => import('@monaco-editor/react'), {
  ssr: false,
})

type MonacoEditorProps = React.ComponentProps<typeof MonacoEditor>

interface ICodeEditor extends MonacoEditorProps {
  wrapperClassName?: string
}

export const CodeEditor = ({
  height = 500,
  options = {
    codeLens: false,
    fontSize: 12,
    fontFamily: '__hack_f5efd2',
    readOnly: true,
    minimap: { enabled: false },
  },
  theme = 'vs-dark',
  wrapperClassName,
  ...props
}: ICodeEditor) => (
  <div
    className={cn('rounded-md min-h-[500px] overflow-hidden', wrapperClassName)}
  >
    <MonacoEditor height={height} theme={theme} options={options} {...props} />
  </div>
)
