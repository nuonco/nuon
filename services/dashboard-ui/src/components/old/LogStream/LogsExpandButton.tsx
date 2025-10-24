'use client'

import React, { type FC } from 'react'
import {
  ArrowsInLineVertical,
  ArrowsOutLineVertical,
} from '@phosphor-icons/react'
import { useLogsViewer } from './logs-viewer-context'
import { Button } from '@/components/old/Button'

export const LogsExpandButton: FC = () => {
  const { isAllExpanded, handleExpandAll } = useLogsViewer()
  return (
    <Button
      className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-2"
      title={isAllExpanded ? 'Collapse all log lines' : 'Expand all log lines'}
      onClick={handleExpandAll}
    >
      {isAllExpanded ? (
        <>
          <ArrowsInLineVertical size="14" /> Collapse
        </>
      ) : (
        <>
          <ArrowsOutLineVertical size="14" /> Expand
        </>
      )}
    </Button>
  )
}
