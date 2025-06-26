'use client'

import React, { useState } from 'react'
import { DateTime } from 'luxon'
import { cn } from '@/stratus/components/helpers'

interface IDateRange {
  start?: string
  end?: string
}

interface DateShortcut {
  label: string
  getValue: () => { start?: string; end?: string }
}

export const defaultShortcuts: DateShortcut[] = [
  {
    label: 'Today',
    getValue: () => ({
      start: DateTime.utc().startOf('day').toISO(),
    }),
  },
  {
    label: 'Yesterday',
    getValue: () => ({
      start: DateTime.utc().minus({ days: 1 }).startOf('day').toISO(),
    }),
  },
  {
    label: 'This week',
    getValue: () => ({
      start: DateTime.utc().startOf('week').toISO(),
      end: DateTime.utc().endOf('week').toISO(),
    }),
  },
  {
    label: 'Last week',
    getValue: () => ({
      start: DateTime.utc().minus({ weeks: 1 }).startOf('week').toISO(),
      end: DateTime.utc().minus({ weeks: 1 }).endOf('week').toISO(),
    }),
  },
  {
    label: 'Last 7 days',
    getValue: () => ({
      start: DateTime.utc().minus({ days: 7 }).startOf('day').toISO(),
      end: DateTime.utc().endOf('day').toISO(),
    }),
  },
  {
    label: 'This month',
    getValue: () => ({
      start: DateTime.utc().startOf('month').toISO(),
      end: DateTime.utc().endOf('month').toISO(),
    }),
  },
  {
    label: 'Last month',
    getValue: () => ({
      start: DateTime.utc().minus({ months: 1 }).startOf('month').toISO(),
      end: DateTime.utc().minus({ months: 1 }).endOf('month').toISO(),
    }),
  },
]

interface IDateRangePicker
  extends Omit<
    React.InputHTMLAttributes<HTMLInputElement>,
    'value' | 'onChange'
  > {
  value?: IDateRange
  onChange?: (value: IDateRange) => void
  isRange?: boolean
  shortcuts?: DateShortcut[]
}

const CalendarMonth = ({
  currentMonth,
  value,
  hoverDate,
  isRange,
  onSelectDate,
}: {
  currentMonth: DateTime
  value?: IDateRange
  hoverDate: DateTime | null
  isRange: boolean
  onSelectDate: (date: DateTime) => void
}) => {
  const firstDayOfMonth = currentMonth.startOf('month')
  const firstDayOfWeek = firstDayOfMonth.weekday % 7

  const days = Array.from({ length: 42 }, (_, i) => {
    const day = firstDayOfMonth.plus({ days: i - firstDayOfWeek })
    const start = value?.start ? DateTime.fromISO(value.start) : null
    const end = value?.end ? DateTime.fromISO(value.end) : null

    return {
      date: day,
      isCurrentMonth: day.month === currentMonth.month,
      isToday: day.toISODate() === DateTime.utc().toISODate(),
      isSelected: isDateSelected(day, start, end, isRange),
      isInRange: isDateInRange(day, start, end, hoverDate),
      isRangeStart: start?.toISODate() === day.toISODate(),
      isRangeEnd: end?.toISODate() === day.toISODate(),
    }
  })

  return (
    <div className="p-4">
      <div className="grid grid-cols-7 gap-1 mb-4">
        {['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su'].map((day) => (
          <div
            key={day}
            className="text-center text-sm font-medium text-gray-400"
          >
            {day}
          </div>
        ))}
      </div>

      <div className="grid grid-cols-7 gap-1">
        {days.map(
          (
            {
              date,
              isCurrentMonth,
              isToday,
              isSelected,
              isInRange,
              isRangeStart,
              isRangeEnd,
            },
            i
          ) => (
            <button
              key={i}
              type="button"
              onClick={() => onSelectDate(date)}
              className={cn(
                'h-8 w-8 rounded-md text-sm flex items-center justify-center relative',
                !isCurrentMonth && 'text-gray-600',
                isCurrentMonth &&
                  !isSelected &&
                  !isInRange &&
                  'text-white hover:bg-gray-700',
                isToday &&
                  !isSelected &&
                  !isInRange &&
                  'border border-purple-500',
                isSelected &&
                  'bg-purple-600 text-white hover:bg-purple-700 z-10',
                isInRange && !isSelected && 'bg-purple-900/50',
                isRangeStart && 'rounded-r-none',
                isRangeEnd && 'rounded-l-none',
                isInRange &&
                  'before:absolute before:inset-y-0 before:left-0 before:right-0 before:bg-purple-900/50 before:-z-10'
              )}
            >
              {date.day}
            </button>
          )
        )}
      </div>
    </div>
  )
}

function isDateSelected(
  date: DateTime,
  start: DateTime | null,
  end: DateTime | null,
  isRange: boolean
) {
  if (!isRange) return start?.toISODate() === date.toISODate()
  return (
    start?.toISODate() === date.toISODate() ||
    end?.toISODate() === date.toISODate()
  )
}

function isDateInRange(
  date: DateTime,
  start: DateTime | null,
  end: DateTime | null,
  hover: DateTime | null
) {
  if (!start) return false
  const endDate = end || hover
  if (!endDate) return false

  const isAfterStart = date >= start
  const isBeforeEnd = date <= endDate
  const isBeforeStart = date <= start
  const isAfterEnd = date >= endDate

  return (isAfterStart && isBeforeEnd) || (isBeforeStart && isAfterEnd)
}

export const DatePicker = ({
  className,
  value,
  onChange,
  isRange = false,
  shortcuts,
  ...props
}: IDateRangePicker) => {
  const [isOpen, setIsOpen] = useState(false)
  const [leftMonth, setLeftMonth] = useState(
    value?.start ? DateTime.fromISO(value.start) : DateTime.utc()
  )
  const [hoverDate, setHoverDate] = useState<DateTime | null>(null)
  const [draftValue, setDraftValue] = useState(value)

  const rightMonth = leftMonth.plus({ months: 1 })

  const handleDateSelect = (date: DateTime) => {
    if (!isRange) {
      setDraftValue({ start: date.toISO() ?? '' })
      return
    }

    if (!draftValue?.start || (draftValue.start && draftValue.end)) {
      setDraftValue({ start: date.toISO() ?? '' })
    } else {
      const start = DateTime.fromISO(draftValue.start)
      const isAfterStart = date > start

      setDraftValue({
        start: isAfterStart ? draftValue.start : (date.toISO() ?? ''),
        end: isAfterStart ? (date.toISO() ?? '') : draftValue.start,
      })
    }
  }

  const handlePrevMonth = () => {
    setLeftMonth(leftMonth.minus({ months: 1 }))
  }

  const handleNextMonth = () => {
    setLeftMonth(leftMonth.plus({ months: 1 }))
  }

  const handleShortcutSelect = (shortcut: DateShortcut) => {
    setDraftValue(shortcut.getValue())
  }

  const handleApply = () => {
    onChange?.(draftValue ?? {})
    setIsOpen(false)
  }

  const handleClear = () => {
    setDraftValue({})
  }

  const formatDateRange = () => {
    if (!value?.start) return ''
    if (!isRange) return DateTime.fromISO(value.start).toFormat('yyyy-MM-dd')
    if (!value.end)
      return `${DateTime.fromISO(value.start).toFormat('yyyy-MM-dd')} to ...`
    return `${DateTime.fromISO(value.start).toFormat('yyyy-MM-dd')} to ${DateTime.fromISO(value.end).toFormat('yyyy-MM-dd')}`
  }

  return (
    <div className="relative">
      <input
        type="text"
        className={cn(
          'w-full px-4 py-2 bg-gray-800 border border-gray-700 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 text-white',
          className
        )}
        value={formatDateRange()}
        onClick={() => setIsOpen(true)}
        readOnly
        {...props}
      />

      {isOpen && (
        <div
          className="absolute z-10 mt-1 bg-gray-900 border border-gray-800 rounded-md shadow-lg text-white"
          onMouseLeave={() => setHoverDate(null)}
        >
          <div className="flex">
            {shortcuts && (
              <div className="w-48 p-4 border-r border-gray-800">
                {shortcuts.map((shortcut) => (
                  <button
                    key={shortcut.label}
                    onClick={() => handleShortcutSelect(shortcut)}
                    className="block w-full text-left px-2 py-2 text-sm hover:bg-gray-800 rounded-md"
                  >
                    {shortcut.label}
                  </button>
                ))}
              </div>
            )}

            <div>
              <div className="p-4 border-b border-gray-800 flex items-center justify-between">
                <button
                  type="button"
                  onClick={handlePrevMonth}
                  className="p-1 hover:bg-gray-800 rounded-md"
                >
                  ←
                </button>
                <div className="flex gap-8">
                  <div className="font-semibold text-center flex items-center gap-2">
                    {leftMonth.toFormat('MMMM yyyy')}
                    <button className="p-1 hover:bg-gray-800 rounded-md">
                      ▾
                    </button>
                  </div>
                  {isRange && (
                    <div className="font-semibold text-center flex items-center gap-2">
                      {rightMonth.toFormat('MMMM yyyy')}
                      <button className="p-1 hover:bg-gray-800 rounded-md">
                        ▾
                      </button>
                    </div>
                  )}
                </div>
                <button
                  type="button"
                  onClick={handleNextMonth}
                  className="p-1 hover:bg-gray-800 rounded-md"
                >
                  →
                </button>
              </div>

              <div className="flex">
                <div onMouseEnter={() => setHoverDate(null)}>
                  <CalendarMonth
                    currentMonth={leftMonth}
                    value={draftValue}
                    hoverDate={hoverDate}
                    isRange={isRange}
                    onSelectDate={handleDateSelect}
                  />
                </div>
                {isRange && (
                  <div
                    className="border-l border-gray-800"
                    onMouseEnter={() => setHoverDate(null)}
                  >
                    <CalendarMonth
                      currentMonth={rightMonth}
                      value={draftValue}
                      hoverDate={hoverDate}
                      isRange={isRange}
                      onSelectDate={handleDateSelect}
                    />
                  </div>
                )}
              </div>

              <div className="p-4 border-t border-gray-800 flex justify-end gap-2">
                <button
                  onClick={handleClear}
                  className="px-4 py-2 text-sm text-gray-300 hover:text-white"
                >
                  Clear
                </button>
                <button
                  onClick={handleApply}
                  className="px-4 py-2 text-sm bg-purple-600 text-white rounded-md hover:bg-purple-700"
                >
                  Apply
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export const ExampleDatePicker = () => {
  const [singleDate, setSingleDate] = useState<{ start?: string }>({
    start: '2025-06-26T14:31:00Z',
  })

  const [dateRange, setDateRange] = useState<{ start?: string; end?: string }>({
    start: '2025-06-26T14:31:00Z',
    end: '2025-06-28T14:31:00Z',
  })

  const [value, setValue] = useState({
    startDate: null,

    endDate: null,
  })

  return (
    <div className="space-y-4 p-4">
      <div>
        <label className="block mb-2 text-sm font-medium text-gray-700">
          Single Date
        </label>
        <DatePicker
          value={singleDate}
          onChange={setSingleDate}
          placeholder="Select a date..."
        />
      </div>

      <div>
        <label className="block mb-2 text-sm font-medium text-gray-700">
          Date Range
        </label>
        <DatePicker
          shortcuts={defaultShortcuts}
          isRange
          value={dateRange}
          onChange={setDateRange}
          placeholder="Select date range..."
        />
      </div>

      <div></div>
    </div>
  )
}
