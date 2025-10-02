'use client'

import React, { type FC } from 'react'
import { ClickToCopyButton } from '@/components/ClickToCopy'
import { Text } from '@/components/Typography'

interface CommandBlockProps {
  command: string
  title?: string
  description?: string | React.ReactElement
  className?: string
}

export const CommandBlock: FC<CommandBlockProps> = ({
  command,
  title,
  description,
  className,
}) => {
  return (
    <div className={`space-y-2 ${className || ''}`}>
      {title && <Text variant="semi-14">{title}</Text>}
      {description && (
        <Text variant="reg-12" className="text-gray-500 dark:text-gray-400">
          {description}
        </Text>
      )}
      <div className="relative rounded-lg p-3 font-mono text-sm bg-gray-100 dark:bg-gray-800">
        <div className="flex items-center justify-between gap-2">
          <code className="text-gray-800 dark:text-gray-200 flex-1 min-w-0">
            {command}
          </code>
          <ClickToCopyButton
            textToCopy={command}
            className="opacity-70 hover:opacity-100 flex-shrink-0"
          />
        </div>
      </div>
    </div>
  )
}
