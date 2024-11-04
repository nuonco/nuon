'use client'

import React, { type FC } from 'react'
import {
  ArrowsInLineVertical,
  ArrowsOutLineVertical,
  MagnifyingGlass,
  Funnel,
  SortAscending,
  SortDescending,
  X,
} from '@phosphor-icons/react'

import { Button } from '@/components/Button'
import { Dropdown } from '@/components/Dropdown'
import { RadioInput } from '@/components/Input'

export interface IRunnerLogsActions {
  columnFilters: any
  columnSort: any
  globalFilter: string
  handleStatusFilter: any
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
  columnSort,
  globalFilter,
  handleGlobalFilter,
  handleStatusFilter,
  handleColumnSort,
  handleExpandAll,
  clearStatusFilter,
  isAllExpanded = false,
  id,
  shouldHideFilter = false,
  shouldShowExpandAll = false,
}) => {
  return (
    <div className="flex items-center gap-4">
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
      >
        <div>
          <label className="relative">
            <MagnifyingGlass className="text-cool-grey-600 dark:text-cool-grey-500 absolute top-0.5 left-2" />
            <input
              className="rounded-md pl-8 pr-3.5 py-1.5 text-base border bg-white dark:bg-dark-grey-100 placeholder:text-cool-grey-600 dark:placeholder:text-cool-grey-500 md:min-w-80"
              type="search"
              placeholder="Search..."
              value={globalFilter}
              onChange={handleGlobalFilter}
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
        <Dropdown
          alignment="right"
          className="text-sm !font-medium !p-2 h-[32px]"
          id="logs-filter"
          text={
            <>
              <Funnel size="14" /> Filter
            </>
          }
        >
          <div>
            <form>
              <RadioInput
                name={`${id}-status-filter`}
                onChange={handleStatusFilter}
                value="Trace"
                labelText="Trace"
              />

              <RadioInput
                name={`${id}-status-filter`}
                onChange={handleStatusFilter}
                value="Debug"
                labelText="Debug"
              />

              <RadioInput
                name={`${id}-status-filter`}
                onChange={handleStatusFilter}
                value="Info"
                labelText="Info"
              />

              <RadioInput
                name={`${id}-status-filter`}
                onChange={handleStatusFilter}
                value="Warn"
                labelText="Warning"
              />

              <RadioInput
                name={`${id}-status-filter`}
                onChange={handleStatusFilter}
                value="Error"
                labelText="Error"
              />

              <RadioInput
                name={`${id}-status-filter`}
                onChange={handleStatusFilter}
                value="Fatal"
                labelText="Fatal"
              />
              <hr />
              <Button
                className="w-full !rounded-t-none !text-sm flex items-center gap-2 pl-4"
                type="reset"
                onClick={clearStatusFilter}
                variant="ghost"
              >
                <X />
                Clear
              </Button>
            </form>
          </div>
        </Dropdown>
      )}
    </div>
  )
}
