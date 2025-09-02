'use client'

import classNames from 'classnames'
import React, { type FC, useMemo, useState } from 'react'
import { createPortal } from 'react-dom'
import { ArrowsOutSimple } from '@phosphor-icons/react'
import { type ColumnDef } from '@tanstack/react-table'
import { LogsControls } from './LogsControls'
import { LogsViewer } from './LogsViewer'
import type { TLogRecord } from './types'
import { Button } from '@/components/Button'
import { LogLineSeverity } from './LogLineSeverity'
import { Modal } from '@/components/Modal'
import { Time } from '@/components/Time'

export interface ILogsModal {
  heading: React.ReactNode
  logs?: Array<TLogRecord>
}

export const LogsModal: FC<ILogsModal> = ({ heading, logs }) => {
  const [isOpen, setIsOpen] = useState<boolean>(false)

  const lineStyle =
    'text-sm font-mono text-cool-grey-600 dark:text-cool-grey-500'

  const columns: Array<ColumnDef<TLogRecord>> = useMemo(
    () => [
      {
        header: 'Severity',
        accessorKey: 'severity_text',
        cell: (props) => (
          <span className={classNames('flex items-center gap-2')}>
            <LogLineSeverity
              severity_number={props.row.original?.severity_number}
            />
            <span className={lineStyle + ' font-semibold uppercase'}>
              {props.getValue<string>() || 'UNKOWN'}
            </span>
          </span>
        ),
        enableColumFilter: true,
        filterFn: 'arrIncludesSome',
      },
      {
        header: 'Date',
        accessorKey: 'timestamp',
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-3 flex items-center gap-1': true,
            })}
          >
            <Time
              className="!text-[10px]"
              time={props.getValue<string>()}
              useMicro
            />
          </span>
        ),
      },
      {
        header: 'Service',
        accessorKey: 'service_name',
        enableColumSort: false,
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-1 flex items-center': true,
            })}
          >
            <span>{props.getValue<string>()}</span>
          </span>
        ),
      },
      {
        header: 'Content',
        accessorKey: 'body',
        cell: (props) => (
          <span
            className={classNames(lineStyle, {
              'col-span-7 flex items-center break-all': true,
            })}
          >
            <span>{props.getValue<string>()}</span>
          </span>
        ),
      },
    ],
    []
  )

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              actions={<LogsControls showLogExpand showLogFilter />}
              className="mx-6 xl:mx-auto xl:w-8xl"
              hasFixedHeight
              heading={heading}
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <LogsViewer
                data={logs}
                columns={columns}
                enableLogFilter
                showLogAttr
              />
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="flex items-center gap-2 text-sm !font-medium"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <ArrowsOutSimple size="14" />
        View all logs
      </Button>
    </>
  )
}
