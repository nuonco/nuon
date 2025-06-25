'use client'

import dynamic from 'next/dynamic'
import React from 'react'
import { cn } from '@/stratus/components/helpers'

const MonacoDiffEditor = dynamic(
  () => import('@monaco-editor/react').then((m) => m.DiffEditor),
  { ssr: false }
)

type MonacoDiffEditorProps = React.ComponentProps<typeof MonacoDiffEditor>

interface IDiffEditor extends MonacoDiffEditorProps {
  wrapperClassName?: string
}

export const DiffEditor = ({
  height = 500,
  options = {
    renderSideBySide: true,
    readOnly: true,
    fontFamily: '__hack_f5efd2',
    fontSize: 12,
  },
  theme = 'vs-dark',
  wrapperClassName,
  ...props
}: IDiffEditor) => {
  return (
    <div
      className={cn(
        'rounded-md min-h-[500px] overflow-hidden',
        wrapperClassName
      )}
    >
      <MonacoDiffEditor
        height={height}
        theme={theme}
        options={options}
        {...props}
      />
    </div>
  )
}
