'use client'

import React, { type FC } from 'react'
import { Funnel } from '@phosphor-icons/react'
import { useLogsViewer } from './logs-viewer-context'
import { LogLineSeverity } from './LogLineSeverity'
import { Button } from '@/components/old/Button'
import { Dropdown } from '@/components/old/Dropdown'
import { CheckboxInput } from '@/components/old/Input'
import { sentanceCase } from '@/utils'

type TLogSeverityText = 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal'
const FILTER_OPTIONS: Array<TLogSeverityText> = [
  'trace',
  'debug',
  'info',
  'warn',
  'error',
  'fatal',
]

function getSeverityNumber(option: TLogSeverityText): number {
  let severityNum: number

  switch (option) {
    case 'trace':
      severityNum = 4
      break
    case 'debug':
      severityNum = 8
      break
    case 'info':
      severityNum = 12
      break
    case 'warn':
      severityNum = 16
      break
    case 'error':
      severityNum = 20
      break
    default:
      severityNum = 24
  }

  return severityNum
}

const groupClasses = {
  trace: ['group/trace', 'group-hover/trace:block'],
  debug: ['group/debug', 'group-hover/debug:block'],
  info: ['group/info', 'group-hover/info:block'],
  warn: ['group/warn', 'group-hover/warn:block'],
  error: ['group/error', 'group-hover/error:block'],
  fatal: ['group/fatal', 'group-hover/fatal:block'],
}

interface ILogsFilterDropdown {}

export const LogsFilterDropdown: FC<ILogsFilterDropdown> = ({}) => {
  const {
    handleStatusFilter,
    handleStatusOnlyFilter,
    clearStatusFilter,
    columnFilters,
  } = useLogsViewer()

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
            {FILTER_OPTIONS.map((opt) => (
              <div className="flex items-center gap-1" key={opt}>
                <CheckboxInput
                  labelClassName="!w-auto !p-1.5  ml-1.5 rounded-sm"
                  name={opt}
                  onChange={handleStatusFilter}
                  checked={columnFilters
                    ?.at(0)
                    ?.value?.includes(sentanceCase(opt))}
                  value={sentanceCase(opt)}
                />
                <Button
                  className={`flex items-center justify-between w-full mr-1.5 py-1 !px-1 ${groupClasses[opt].at(0)}`}
                  variant="ghost"
                  type="button"
                  value={sentanceCase(opt)}
                  onClick={
                    columnFilters?.at(0)?.value?.length === 1 &&
                    columnFilters?.at(0)?.value?.includes(sentanceCase(opt))
                      ? clearStatusFilter
                      : handleStatusOnlyFilter
                  }
                >
                  <span className="flex items-center gap-1">
                    <LogLineSeverity severity_number={getSeverityNumber(opt)} />
                    <span className="font-semibold font-mono uppercase text-sm">
                      {sentanceCase(opt)}
                    </span>
                  </span>
                  <span
                    className={`text-sm self-end hidden ${groupClasses[opt].at(1)}`}
                  >
                    {columnFilters?.at(0)?.value?.length === 1 &&
                    columnFilters?.at(0)?.value?.includes(sentanceCase(opt))
                      ? 'Reset'
                      : 'Only'}
                  </span>
                </Button>
              </div>
            ))}
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
