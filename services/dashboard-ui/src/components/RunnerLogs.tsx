'use client'

import classNames from 'classnames'
import React, { type FC, useState } from 'react'
import {
  ArrowsOutSimple,
  MagnifyingGlass,
  FunnelSimple,
  Funnel,
} from '@phosphor-icons/react'
import { Button, Text, Time, Section, Modal } from '@/components'
import type { TOTELLog } from '@/types'

const LogLineSeverity: FC<{ severity_number: number }> = ({
  severity_number,
}) => {
  return (
    <span
      className={classNames('flex w-0.5 h-3', {
        'bg-primary-400 dark:bg-primary-300': severity_number <= 4,
        'bg-cool-grey-600 dark:bg-cool-grey-500':
          severity_number >= 5 && severity_number <= 8,
        'bg-blue-600 dark:bg-blue-500':
          severity_number >= 9 && severity_number <= 12,
        'bg-orange-600 dark:bg-orange-500':
          severity_number >= 13 && severity_number <= 16,
        'bg-red-600 dark:bg-red-500':
          severity_number >= 17 && severity_number <= 20,
        'bg-red-700 dark:bg-red-600':
          severity_number >= 21 && severity_number <= 24,
      })}
    />
  )
}

const LogLine: FC<{ line: TOTELLog; isPreview?: boolean }> = ({
  line,
  isPreview = false,
}) => {
  const lineStyle =
    'tracking-wider text-sm font-mono leading-loose text-cool-grey-600 dark:text-cool-grey-500'

  return (
    <span className="grid grid-cols-12 items-center justify-start gap-6 py-2 w-full">
      {isPreview ? null : (
        <span className={classNames('flex items-center gap-2')}>
          <LogLineSeverity severity_number={line.severity_number} />
          <span className={lineStyle + ' font-semibold uppercase'}>
            {line?.severity_text || 'UNKOWN'}
          </span>
        </span>
      )}

      <span
        className={classNames(lineStyle, {
          'col-span-2': !isPreview,
          'col-span-3 flex items-center gap-2': isPreview,
        })}
      >
        {isPreview && (
          <LogLineSeverity severity_number={line.severity_number} />
        )}
        <Time className="!text-sm" time={line.timestamp} />
      </span>
      <span
        className={classNames(lineStyle, {
          'col-span-2': !isPreview,
          'col-span-3': isPreview,
        })}
      >
        {line?.resource_attributes?.['service.name']}
      </span>

      <span
        className={classNames(lineStyle, {
          'col-span-7': !isPreview,
          'col-span-5 truncate': isPreview,
        })}
      >
        {line?.body}
      </span>
    </span>
  )
}

export const OTELLogs: FC<{ logs?: Array<TOTELLog> }> = ({ logs = [] }) => {
  console.log('logs', logs)

  return (
    <div className="divide-y">
      <div className="grid grid-cols-12 items-center justify-start gap-6 py-2 w-full">
        <Text className="!font-medium text-cool-grey-600 dark:text-cool-grey-500">
          Severity
        </Text>
        <Text className="!font-medium text-cool-grey-600 dark:text-cool-grey-500 col-span-2">
          Date
        </Text>
        <Text className="!font-medium text-cool-grey-600 dark:text-cool-grey-500 col-span-2">
          Service
        </Text>
        <Text className="!font-medium text-cool-grey-600 dark:text-cool-grey-500 col-span-7">
          Content
        </Text>
      </div>
      {logs.map((line) => (
        <LogLine key={line?.timestamp as string} line={line} />
      ))}
    </div>
  )
}

export interface ILogsPreview {
  logs: Array<TOTELLog>
}

export const LogsPreview: FC<ILogsPreview> = ({ logs }) => {
  return (
    <div className="divide-y">
      <div className="grid grid-cols-12 items-center justify-start gap-6 py-2 w-full">
        <Text className="!font-medium text-cool-grey-600 dark:text-cool-grey-500 col-span-3">
          Date
        </Text>
        <Text className="!font-medium text-cool-grey-600 dark:text-cool-grey-500 col-span-3">
          Service
        </Text>
        <Text className="!font-medium text-cool-grey-600 dark:text-cool-grey-500 col-span-5">
          Content
        </Text>
      </div>
      {logs.slice(0, 15).map((line) => (
        <LogLine key={line?.timestamp as string} line={line} isPreview />
      ))}
    </div>
  )
}

export interface IRunnerLogs {
  heading: React.ReactNode
  logs: Array<TOTELLog>
}

export const RunnerLogs: FC<IRunnerLogs> = ({ heading, logs }) => {
  const [isDetailsOpen, setIsDetailsOpen] = useState<boolean>(false)

  return (
    <>
      <Modal
        actions={<RunnerLogsActions />}
        heading={heading}
        isOpen={isDetailsOpen}
        onClose={() => {
          setIsDetailsOpen(false)
        }}
      >
        <OTELLogs logs={logs} />
      </Modal>
      <Section
        className="border-r"
        actions={
          <div className="flex items-center divide-x">
            <div className="pr-4">
              <RunnerLogsActions />
            </div>
            <div className="pl-4">
              <Button
                className="flex items-center gap-2 text-base !font-medium"
                onClick={() => {
                  setIsDetailsOpen(true)
                }}
              >
                <ArrowsOutSimple />
                Open logs
              </Button>
            </div>
          </div>
        }
        heading={heading}
      >
        {logs?.length ? ( <LogsPreview logs={logs} />) : <Text className="text-base">No logs found</Text>}
      </Section>
    </>
  )
}

const RunnerLogsActions: FC = () => {
  return (
    <div className="flex items-center gap-4">
      <Button
        className="text-base !font-medium !p-2 w-[32px] h-[32px]"
        variant="ghost"
      >
        <MagnifyingGlass />
      </Button>

      <Button
        className="text-base !font-medium !p-2 w-[32px] h-[32px]"
        variant="ghost"
      >
        <FunnelSimple />
      </Button>

      <Button
        className="text-base !font-medium !p-2 w-[32px] h-[32px]"
        variant="ghost"
      >
        <Funnel />
      </Button>
    </div>
  )
}
