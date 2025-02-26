'use client'

import React, { type FC } from 'react'
import {
  ArrowClockwise,
  ArrowsInLineVertical,
  ArrowsOutLineVertical,
  MagnifyingGlass,
  Funnel,
  SortAscending,
  SortDescending,
} from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Dropdown } from '@/components/Dropdown'
import { CheckboxInput, Input } from '@/components/Input'
import { LogLineSeverity } from '@/components/RunnerLogLineSeverity'
import { RunnerLogFilterDropdown } from '@/components/RunnerLogFilterDropdown'

export interface IRunnerLogsActions {
  columnFilters: any
  columnSort: any
  globalFilter: string
  handleStatusFilter: any
  handleStatusOnlyFilter: any
  handleGlobalFilter: any
  handleColumnSort: any
  handleExpandAll?: any
  clearStatusFilter?: any
  isAllExpanded?: boolean
  id: string
  shouldHideFilter?: boolean
  shouldShowExpandAll?: boolean
}

export const RunnerLogsActions: FC<IRunnerLogsActions> = ({
  columnFilters,
  columnSort,
  globalFilter,
  handleGlobalFilter,
  handleStatusFilter,
  handleStatusOnlyFilter,
  handleColumnSort,
  handleExpandAll,
  clearStatusFilter,
  isAllExpanded = false,
  id,
  shouldHideFilter = false,
  shouldShowExpandAll = false,
}) => {
  return (
    <div className="flex items-center gap-4 pl-4">
      {shouldShowExpandAll && (
        <Button
          className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-2"
          title={
            isAllExpanded ? 'Collapse all log lines' : 'Expand all log lines'
          }
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
      )}
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
      {shouldHideFilter ? null : (
        <RunnerLogFilterDropdown
          clearStatusFilter={clearStatusFilter}
          columnFilters={columnFilters}
          handleStatusFilter={handleStatusFilter}
          handleStatusOnlyFilter={handleStatusOnlyFilter}
        />
      )}
    </div>
  )
}
