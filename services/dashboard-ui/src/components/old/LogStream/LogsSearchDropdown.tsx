'use client'

import React, { type FC } from 'react'
import { MagnifyingGlass, XCircle } from '@phosphor-icons/react'
import { useLogsViewer } from './logs-viewer-context'
import { Button } from '@/components/old/Button'
import { Dropdown } from '@/components/old/Dropdown'
import { Input } from '@/components/old/Input'

export interface ILogsSearchDropdown {}

export const LogsSearchDropdown: FC<ILogsSearchDropdown> = ({}) => {
  const { globalFilter, handleGlobalFilter } = useLogsViewer()

  return (
    <Dropdown
      alignment="right"
      className="text-sm !font-medium !p-2 h-[32px]"
      id="logs-search"
      text={
        <>
          <MagnifyingGlass size="14" /> Search
        </>
      }
      isDownIcon
    >
      <div className="p-2">
        <label className="relative">
          <MagnifyingGlass className="text-cool-grey-600 dark:text-cool-grey-500 absolute top-0.5 left-2" />
          <Input
            className="md:min-w-80"
            type="search"
            placeholder="Search..."
            value={globalFilter}
            onChange={handleGlobalFilter}
            isSearch
          />
          {globalFilter !== '' ? (
            <Button
              className="!p-0.5 absolute top-1/2 right-1.5 -translate-y-1/2"
              variant="ghost"
              title="clear search"
              value=""
              onClick={handleGlobalFilter}
            >
              <XCircle />
            </Button>
          ) : null}
        </label>
      </div>
    </Dropdown>
  )
}
