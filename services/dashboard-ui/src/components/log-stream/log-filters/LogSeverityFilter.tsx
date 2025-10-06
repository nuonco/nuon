'use client'

import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { useLogs } from '@/hooks/use-logs'
import { cn } from '@/utils/classnames'
import { getSeverityBorderClasses } from '@/utils/log-stream-utils'

type TLogSeverityText = 'Trace' | 'Debug' | 'Info' | 'Warn' | 'Error' | 'Fatal'
const LOG_ACTIONS: {
  value: number
  label: TLogSeverityText
}[] = [
  { value: 4, label: 'Trace' },
  { value: 8, label: 'Debug' },
  { value: 12, label: 'Info' },
  { value: 16, label: 'Warn' },
  { value: 20, label: 'Error' },
  { value: 24, label: 'Fatal' },
]

interface ILogSeverityFilter {
  title: string
}

export const LogSeverityFilter = ({ title }: ILogSeverityFilter) => {
  const {
    filters: {
      selectedSeverities,
      handleSeverityInputToggle,
      handleSeverityButtonClick,
      handleSeverityReset,
    },
  } = useLogs()
  const getButtonText = (action: string) => {
    if (selectedSeverities.size === 1 && selectedSeverities.has(action)) {
      return 'Reset'
    }
    return 'Only'
  }

  const actionOptions = LOG_ACTIONS

  return (
    <Dropdown
      buttonClassName="!p-1"
      buttonText={`Filter ${title} (${selectedSeverities.size})`}
      className="ml-auto"
      icon={<Icon variant="FadersHorizontal" />}
      id={`${title}-fitler`}
      alignment="right"
      variant="ghost"
    >
      <Menu className="min-w-56">
        {actionOptions.map((option) => (
          <div key={option.value} className="flex items-center space-x-2">
            <input
              type="checkbox"
              id={`${title.toLowerCase().replace(' ', '-')}-action-${option.label}`}
              checked={selectedSeverities.has(option.label)}
              onChange={() => handleSeverityInputToggle(option.label)}
              className="focus:ring-primary-500 focus:border-primary-500 accent-primary-600"
            />
            <Button
              className="!p-1 flex items-center justify-between group w-full"
              variant="ghost"
              size="sm"
              onClick={() => handleSeverityButtonClick(option.label)}
            >
              <span className="flex items-center gap-1.5 h-full">
                <span
                  className={cn('inline-block border-l-2 h-full', {
                    [getSeverityBorderClasses(option.value, 'l')]: true,
                  })}
                />
                {option.label}
              </span>
              <span className="hidden group-hover:inline-flex text-xs">
                {getButtonText(option.label)}
              </span>
            </Button>
          </div>
        ))}

        <hr />
        <Button
          className="w-full !p-1 shrink-0"
          variant="ghost"
          size="sm"
          onClick={handleSeverityReset}
        >
          Reset
        </Button>
      </Menu>
    </Dropdown>
  )
}
