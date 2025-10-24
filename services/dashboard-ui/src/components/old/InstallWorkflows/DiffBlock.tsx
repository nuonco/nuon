'use client'

import React from 'react'
import { CodeBlock } from '../CodeBlock'
import { Text } from '../Typography'

interface DiffBlockProps {
  className?: string
  children?: string
}

export const DiffBlock: React.FC<DiffBlockProps> = ({
  className,
  children,
}) => {
  // Removed console.log to fix linting error

  if (!children) {
    return (
      <div
        className={`p-4 bg-cool-grey-50 dark:bg-dark-grey-300 rounded ${className || ''}`}
      >
        <Text variant="reg-14" isMuted>
          No diff content available.
        </Text>
      </div>
    )
  }

  // Check if this is an "all unchanged" diff - all lines start with two spaces
  const lines = children.split('\n')
  const allUnchanged = lines.every((line) => line.startsWith('  '))
  const hasChanges = lines.some(
    (line) => line.startsWith('+ ') || line.startsWith('- ')
  )

  // If it's just a display of a resource with no changes
  if (allUnchanged && !hasChanges) {
    return (
      <div className={`${className || ''}`}>
        <div className="mb-2 px-3 py-2 bg-blue-50 dark:bg-blue-900 rounded-md">
          <Text variant="reg-14" className="text-blue-700 dark:text-blue-300">
            This resource has no content changes
          </Text>
        </div>
        <CodeBlock language="yaml">
          {lines.map((line) => line.substring(2)).join('\n')}
        </CodeBlock>
      </div>
    )
  }

  // If there are actual changes (additions/removals)
  return (
    <div className={`${className || ''}`}>
      <CodeBlock language="diff" isDiff={true}>
        {children}
      </CodeBlock>
    </div>
  )
}
