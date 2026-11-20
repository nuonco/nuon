import { Text } from '@/components/common/Text'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'
import { cn } from '@/utils/classnames'
import {
  HEALTH_DEGRADED,
  HEALTH_HEALTHY,
  HEALTH_PROGRESSING,
  HEALTH_UNHEALTHY,
  HEALTH_UNKNOWN,
} from '@/utils/health-utils'
import {
  HEALTH_FILTER_OPTIONS,
  healthFilterLabel,
  NO_SIGNAL,
} from './fleet-utils'
import { FleetFilterDropdown } from './FleetFilterDropdown'

const HEALTH_BORDER_CLASSES: Record<string, string> = {
  [HEALTH_UNHEALTHY]: '!border-l-red-600 dark:!border-l-red-500',
  [HEALTH_DEGRADED]: '!border-l-orange-600 dark:!border-l-orange-500',
  [HEALTH_PROGRESSING]: '!border-l-blue-600 dark:!border-l-blue-500',
  [HEALTH_HEALTHY]: '!border-l-green-600 dark:!border-l-green-500',
  [HEALTH_UNKNOWN]: '!border-l-cool-grey-600 dark:!border-l-cool-grey-500',
  [NO_SIGNAL]: '!border-l-cool-grey-600 dark:!border-l-cool-grey-500',
}

export interface IHealthFilterDropdown {
  selected: Set<string>
  onChange: (selected: Set<string>) => void
  id?: string
  label?: string
  options?: string[]
  optionLabel?: (value: string) => string
}

export const HealthFilterDropdown = ({
  selected,
  onChange,
  id = 'fleet-filter-health',
  label = 'Health',
  options = HEALTH_FILTER_OPTIONS,
  optionLabel = healthFilterLabel,
}: IHealthFilterDropdown) => {
  const toggle = (value: string) => {
    const next = new Set(selected)
    if (next.has(value)) next.delete(value)
    else next.add(value)
    onChange(next)
  }

  const only = (value: string) =>
    onChange(
      selected.size === 1 && selected.has(value) ? new Set() : new Set([value])
    )

  return (
    <FleetFilterDropdown
      id={id}
      label={label}
      count={selected.size}
      onReset={() => onChange(new Set())}
    >
      {options.map((value) => {
        const checked = selected.has(value)
        const isOnlySelected = selected.size === 1 && checked
        return (
          <CheckboxInputWithButton
            key={value}
            className={cn(
              'w-full border-l-4 pl-2 pr-0',
              HEALTH_BORDER_CLASSES[value]
            )}
            name={`${id}-${value}`}
            value={value}
            checked={checked}
            onChange={() => toggle(value)}
            buttonProps={{
              className:
                '!p-1 flex items-center justify-between group w-full !rounded-sm',
              variant: 'ghost',
              type: 'button',
              value,
              onClick: () => only(value),
              children: (
                <>
                  <Text variant="body" weight="strong">
                    {optionLabel(value)}
                  </Text>
                  <Text
                    variant="subtext"
                    theme="neutral"
                    className="ml-2 opacity-0 group-hover:opacity-100"
                  >
                    {isOnlySelected ? 'Reset' : 'Only'}
                  </Text>
                </>
              ),
            }}
          />
        )
      })}
    </FleetFilterDropdown>
  )
}
