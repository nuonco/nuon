'use client'

import React, { type FC } from 'react'
import { useLogs } from './logs-context'
import { LogsExpandButton } from './LogsExpandButton'
import { LogsFilterDropdown } from './LogsFilterDropdown'
import { LogsSearchDropdown } from './LogsSearchDropdown'
import { LogsSortButton } from './LogsSortButton'

export interface ILogsControls {
  showLogSearch?: boolean
  showLogSort?: boolean
  showLogFilter?: boolean
  showLogExpand?: boolean
}

export const LogsControls: FC<ILogsControls> = ({
  showLogExpand = false,
  showLogFilter = false,
  showLogSearch = true,
  showLogSort = true,
}) => {
  const { error, logs } = useLogs()

  return !error && logs.length ? (
    <div className="flex items-center gap-4">
      {showLogExpand ? <LogsExpandButton /> : null}
      {showLogSearch ? <LogsSearchDropdown /> : null}
      {showLogSort ? <LogsSortButton /> : null}
      {showLogFilter ? <LogsFilterDropdown /> : null}
    </div>
  ) : null
}
