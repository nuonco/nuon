'use client'

import React, { type FC, createContext, useContext, useState } from 'react'

interface ILogsViewerContext {
  columnFilters?: any
  globalFilter?: any
  columnSort?: any
  isAllExpanded?: boolean
  handleStatusFilter?: any
  handleStatusOnlyFilter?: any
  clearStatusFilter?: any
  handleGlobalFilter?: any
  handleColumnSort?: any
  handleExpandAll?: any
}

const LogsViewerContext = createContext<ILogsViewerContext>({})

export interface ILogsViewerProvider {
  children: React.ReactNode
}

export const LogsViewerProvider: FC<ILogsViewerProvider> = ({
  children,
}) => {
  const [columnFilters, setColumnFilters] = useState([
    {
      id: 'severity_text',
      value: ['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal'],
    },
  ])
  const [globalFilter, setGlobalFilter] = useState('')
  const [columnSort, setColumnSort] = useState([
    { id: 'timestamp', desc: true },
  ])
  const [isAllExpanded, setIsAllExpanded] = useState(false)

  const handleStatusFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { checked, value } = e.target
    setColumnFilters((state) => {
      const values = [...state?.at(0)?.value]
      const index = values?.indexOf(value)
      if (checked && index < 0) {
        values.push(value)
      } else if (index > -1) {
        values.splice(index, 1)
      }
      return [{ id: 'severity_text', value: values }]
    })
  }

  const handleStatusOnlyFilter = (e: React.MouseEvent<HTMLButtonElement>) => {
    setColumnFilters([
      { id: 'severity_text', value: [e?.currentTarget?.value] },
    ])
  }

  const clearStatusFilter = () => {
    setColumnFilters([
      {
        id: 'severity_text',
        value: ['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal'],
      },
    ])
  }

  const handleGlobalFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGlobalFilter(e.target.value || "")
  }

  const handleColumnSort = () => {
    setColumnSort([{ id: 'timestamp', desc: !columnSort?.[0].desc }])
  }

  const handleExpandAll = () => {
    setIsAllExpanded(!isAllExpanded)
  }

  return (
    <LogsViewerContext.Provider
      value={{
        clearStatusFilter,
        columnFilters,
        columnSort,
        globalFilter,
        handleColumnSort,
        handleExpandAll,
        handleGlobalFilter,
        handleStatusFilter,
        handleStatusOnlyFilter,
        isAllExpanded,
      }}
    >
      {children}
    </LogsViewerContext.Provider>
  )
}

export const useLogsViewer = (): ILogsViewerContext => {
  return useContext(LogsViewerContext)
}
