'use client'

import dynamic from 'next/dynamic'
import React from 'react'

// Dynamically import the Monaco Editor to disable SSR
const MonacoEditor = dynamic(() => import('@monaco-editor/react'), {
  ssr: false,
})

type Props = {
  value: string
  language?: string
  onChange?: (value: string | undefined) => void
  height?: number | string
}

export const Editor: React.FC<Props> = ({
  value,
  language = 'javascript',
  height = 400,
  onChange,
}) => (
  <MonacoEditor
    height={height}
    defaultLanguage={language}
    defaultValue={value}
    onChange={onChange}
    theme="vs-dark"
    options={{
      readOnly: true,
      fontSize: 12,
      minimap: { enabled: false },
    }}
  />
)

function splitDiffToYaml(diffText) {
  const originalLines = []
  const modifiedLines = []

  diffText.split('\n').forEach((line) => {
    if (line.startsWith('-')) {
      originalLines.push(line.slice(1).trimStart())
    } else if (line.startsWith('+')) {
      modifiedLines.push(line.slice(1).trimStart())
    } else if (line.trim() !== '') {
      // Unchanged line: add to both
      originalLines.push(line)
      modifiedLines.push(line)
    }
    // Ignore empty lines at start/end
  })

  return {
    original: originalLines.join('\n'),
    modified: modifiedLines.join('\n'),
  }
}

const MonacoDiffEditor = dynamic(
  () => import('@monaco-editor/react').then((m) => m.DiffEditor),
  { ssr: false }
)

export const DiffEditor = ({ diff }: { diff: string }) => {
  const splitDiff = splitDiffToYaml(diff)
  return (
    <MonacoDiffEditor
      original={splitDiff?.original}
      modified={splitDiff?.modified}
      language="yaml"
      height={500}
      theme="vs-dark"
      options={{
        renderSideBySide: true, // true for side-by-side, false for inline
        readOnly: true,
      }}
    />
  )
}
