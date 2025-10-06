'use client'

import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { useLogs } from '@/hooks/use-logs'

type TLogServiceText = 'api' | 'runner'
const LOG_ACTIONS: {
  value: number
  label: TLogServiceText
}[] = [
  { value: 4, label: 'api' },
  { value: 8, label: 'runner' },
]

interface ILogServiceFilter {
  title: string
}

export const LogServiceFilter = ({ title }: ILogServiceFilter) => {
  const {
    filters: {
      selectedServices,
      handleServiceInputToggle,
      handleServiceButtonClick,
      handleServiceReset,
    },
  } = useLogs()

  const getButtonText = (action: string) => {
    if (selectedServices.size === 1 && selectedServices.has(action)) {
      return 'Reset'
    }
    return 'Only'
  }

  const actionOptions = LOG_ACTIONS

  return (
    <Dropdown
      buttonClassName="!p-1"
      buttonText={`Filter ${title} (${selectedServices.size})`}
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
              checked={selectedServices.has(option.label)}
              onChange={() => handleServiceInputToggle(option.label)}
              className="focus:ring-primary-500 focus:border-primary-500 accent-primary-600"
            />
            <Button
              className="!p-1 flex items-center justify-between group w-full"
              variant="ghost"
              size="sm"
              onClick={() => handleServiceButtonClick(option.label)}
            >
              <span className="flex items-center gap-1.5 h-full">
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
          onClick={handleServiceReset}
        >
          Reset
        </Button>
      </Menu>
    </Dropdown>
  )
}
