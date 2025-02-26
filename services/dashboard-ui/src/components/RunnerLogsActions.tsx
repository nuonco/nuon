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
  X,
} from '@phosphor-icons/react'

import { Button } from '@/components/Button'
import { Dropdown } from '@/components/Dropdown'
import { CheckboxInput, Input } from '@/components/Input'
import { LogLineSeverity } from '@/components/RunnerLogLineSeverity'

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
  columnFilters,
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
        <Dropdown
          alignment="right"
          className="text-sm !font-medium !p-2 h-[32px]"
          id="logs-filter"
          text={
            <>
              <Funnel size="14" /> Filter
            </>
          }
          isDownIcon
        >
          <div>
            <form>
              <CheckboxInput
                name="trace"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Trace')}
                value="Trace"
                labelText={
                  <div className="flex items-center gap-1">
                    <LogLineSeverity severity_number={4} />
                    <span className="font-semibold font-mono uppercase">
                      Trace
                    </span>
                  </div>
                }
              />

              <CheckboxInput
                name="debug"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Debug')}
                value="Debug"
                labelText={
                  <div className="flex items-center gap-1">
                    <LogLineSeverity severity_number={8} />
                    <span className="font-semibold !font-mono uppercase">
                      Debug
                    </span>
                  </div>
                }
              />

              <CheckboxInput
                name="info"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Info')}
                value="Info"
                labelText={
                  <div className="flex items-center gap-1">
                    <LogLineSeverity severity_number={12} />
                    <span className="font-semibold !font-mono uppercase">
                      Info
                    </span>
                  </div>
                }
              />

              <CheckboxInput
                name="warn"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Warn')}
                value="Warn"
                labelText={
                  <div className="flex items-center gap-1">
                    <LogLineSeverity severity_number={16} />
                    <span className="font-semibold !font-mono uppercase">
                      Warn
                    </span>
                  </div>
                }
              />

              <CheckboxInput
                name="error"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Error')}
                value="Error"
                labelText={
                  <div className="flex items-center gap-1">
                    <LogLineSeverity severity_number={20} />
                    <span className="font-semibold !font-mono uppercase">
                      Error
                    </span>
                  </div>
                }
              />

              <CheckboxInput
                name="fatal"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Fatal')}
                value="Fatal"
                labelText={
                  <div className="flex items-center gap-1">
                    <LogLineSeverity severity_number={24} />
                    <span className="font-semibold !font-mono uppercase">
                      Fatal
                    </span>
                  </div>
                }
              />
              <hr />
              <Button
                className="w-full !rounded-t-none !text-sm flex items-center gap-2 pl-4"
                type="button"
                onClick={clearStatusFilter}
                variant="ghost"
              >
                <ArrowClockwise />
                Reset
              </Button>
            </form>
          </div>
        </Dropdown>
      )}
    </div>
  )
}
