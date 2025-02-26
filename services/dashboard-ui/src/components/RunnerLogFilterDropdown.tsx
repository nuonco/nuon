'use client'

import React, { type FC } from 'react'
import { ArrowClockwise, Funnel } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Dropdown } from '@/components/Dropdown'
import { CheckboxInput } from '@/components/Input'
import { LogLineSeverity } from '@/components/RunnerLogLineSeverity'

interface IRunnerLogFilterDropdown {
  handleStatusFilter: any
  handleStatusOnlyFilter: any
  columnFilters: any
  clearStatusFilter: any
}

export const RunnerLogFilterDropdown: FC<IRunnerLogFilterDropdown> = ({
  handleStatusFilter,
  handleStatusOnlyFilter,
  clearStatusFilter,
  columnFilters,
}) => {
  return (
    <Dropdown
      alignment="right"
      className="text-sm !font-medium !p-2 h-[32px]"
      id="logs-filter"
      text={
        <>
          <Funnel size="14" /> Filter ({columnFilters?.at(0)?.value?.length})
        </>
      }
      isDownIcon
    >
      <div className="min-w-[200px]">
        <form>
          <div className="py-2 flex flex-col gap-1">
            <div className="flex items-center gap-1">
              <CheckboxInput
                labelClassName="!w-auto !p-1.5  ml-1.5 rounded-sm"
                name="trace"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Trace')}
                value="Trace"
              />
              <Button
                className="flex items-center justify-between w-full mr-1.5 py-1 !px-1 group/trace"
                variant="ghost"
                type="button"
                value="Trace"
                onClick={handleStatusOnlyFilter}
              >
                <span className="flex items-center gap-1">
                  <LogLineSeverity severity_number={4} />
                  <span className="font-semibold font-mono uppercase text-sm">
                    Trace
                  </span>
                </span>
                <span className="text-sm self-end hidden group-hover/trace:block">
                  Only
                </span>
              </Button>
            </div>

            <div className="flex items-center gap-1">
              <CheckboxInput
                labelClassName="!w-auto !p-1.5  ml-1.5 rounded-sm"
                name="debug"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Debug')}
                value="Debug"
              />
              <Button
                className="flex items-center justify-between w-full mr-1.5 py-1 !px-1 group/debug"
                variant="ghost"
                type="button"
                value="Debug"
                onClick={handleStatusOnlyFilter}
              >
                <span className="flex items-center gap-1">
                  <LogLineSeverity severity_number={8} />
                  <span className="font-semibold font-mono uppercase text-sm">
                    Debug
                  </span>
                </span>
                <span className="text-sm self-end hidden group-hover/debug:block">
                  Only
                </span>
              </Button>
            </div>

            <div className="flex items-center gap-1">
              <CheckboxInput
                labelClassName="!w-auto !p-1.5  ml-1.5 rounded-sm"
                name="info"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Info')}
                value="Info"
              />
              <Button
                className="flex items-center justify-between w-full mr-1.5 py-1 !px-1 group/info"
                variant="ghost"
                type="button"
                value="Info"
                onClick={handleStatusOnlyFilter}
              >
                <span className="flex items-center gap-1">
                  <LogLineSeverity severity_number={12} />
                  <span className="font-semibold font-mono uppercase text-sm">
                    Info
                  </span>
                </span>
                <span className="text-sm self-end hidden group-hover/info:block">
                  Only
                </span>
              </Button>
            </div>

            <div className="flex items-center gap-1">
              <CheckboxInput
                labelClassName="!w-auto !p-1.5  ml-1.5 rounded-sm"
                name="warn"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Warn')}
                value="Warn"
              />
              <Button
                className="flex items-center justify-between w-full mr-1.5 py-1 !px-1 group/warn"
                variant="ghost"
                type="button"
                value="Warn"
                onClick={handleStatusOnlyFilter}
              >
                <span className="flex items-center gap-1">
                  <LogLineSeverity severity_number={16} />
                  <span className="font-semibold font-mono uppercase text-sm">
                    Warn
                  </span>
                </span>
                <span className="text-sm self-end hidden group-hover/warn:block">
                  Only
                </span>
              </Button>
            </div>

            <div className="flex items-center gap-1">
              <CheckboxInput
                labelClassName="!w-auto !p-1.5  ml-1.5 rounded-sm"
                name="error"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Error')}
                value="Error"
              />
              <Button
                className="flex items-center justify-between w-full mr-1.5 py-1 !px-1 group/error"
                variant="ghost"
                type="button"
                value="Error"
                onClick={handleStatusOnlyFilter}
              >
                <span className="flex items-center gap-1">
                  <LogLineSeverity severity_number={20} />
                  <span className="font-semibold font-mono uppercase text-sm">
                    Error
                  </span>
                </span>
                <span className="text-sm self-end hidden group-hover/error:block">
                  Only
                </span>
              </Button>
            </div>

            <div className="flex items-center gap-1">
              <CheckboxInput
                labelClassName="!w-auto !p-1.5  ml-1.5 rounded-sm"
                name="fatal"
                onChange={handleStatusFilter}
                checked={columnFilters?.at(0)?.value?.includes('Fatal')}
                value="Fatal"
              />
              <Button
                className="flex items-center justify-between w-full mr-1.5 py-1 !px-1 group/fatal"
                variant="ghost"
                type="button"
                value="Fatal"
                onClick={handleStatusOnlyFilter}
              >
                <span className="flex items-center gap-1">
                  <LogLineSeverity severity_number={24} />
                  <span className="font-semibold font-mono uppercase text-sm">
                    Fatal
                  </span>
                </span>
                <span className="text-sm self-end hidden group-hover/fatal:block">
                  Only
                </span>
              </Button>
            </div>
          </div>

          <hr />
          <Button
            className="w-full !rounded-t-none !text-base flex items-center gap-2 pl-4"
            type="button"
            onClick={clearStatusFilter}
            variant="ghost"
          >
            Reset
          </Button>
        </form>
      </div>
    </Dropdown>
  )
}
