'use client'

import React, { type FC } from 'react'
import { MagnifyingGlass } from '@phosphor-icons/react'
import { useLogsViewer } from './logs-viewer-context'
import { Dropdown } from '@/components/Dropdown'
import { Input } from '@/components/Input'

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
        </label>
      </div>
    </Dropdown>
  )
}
