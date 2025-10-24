'use client'

import React, { type FC } from 'react'
import { SortAscending, SortDescending } from '@phosphor-icons/react'
import { useLogsViewer } from './logs-viewer-context'
import { Button } from '@/components/old/Button'

export const LogsSortButton: FC = () => {
  const { columnSort, handleColumnSort } = useLogsViewer()
  return (
    <Button
      className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-2"
      title={columnSort?.[0].desc ? 'Sort by oldest' : 'Sort by newest'}
      onClick={handleColumnSort}
    >
      <>
        {columnSort?.[0].desc ? (
          <SortAscending size="14" />
        ) : (
          <SortDescending size="14" />
        )}
        Sort
      </>
    </Button>
  )
}
