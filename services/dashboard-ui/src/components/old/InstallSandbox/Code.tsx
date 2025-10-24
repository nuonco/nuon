'use client'

import React, { type FC } from 'react'
import { ClickToCopyButton } from '@/components/old/ClickToCopy'
import { Text } from '@/components/old/Typography'
import { CodeViewer } from '@/components/old/Code'

export const Code: FC<any> = ({ children }) => {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between">
        <Text variant="med-12">Nuon Terraform backend config</Text>
        <ClickToCopyButton
          textToCopy={children.slice(1, children.length - 1)}
        />
      </div>
      <CodeViewer initCodeSource={children} language="hcl" />
    </div>
  )
}
