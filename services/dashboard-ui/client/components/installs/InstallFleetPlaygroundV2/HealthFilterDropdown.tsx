import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
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
import { humanize } from '@/utils/string-utils'
import { HEALTH_FILTER_OPTIONS, NO_SIGNAL } from './fleet'

const HEALTH_BORDER_CLASSES: Record<string, string> = {
  [HEALTH_UNHEALTHY]: '!border-l-red-600 dark:!border-l-red-500',
  [HEALTH_DEGRADED]: '!border-l-orange-600 dark:!border-l-orange-500',
  [HEALTH_PROGRESSING]: '!border-l-blue-600 dark:!border-l-blue-500',
  [HEALTH_HEALTHY]: '!border-l-green-600 dark:!border-l-green-500',
  [HEALTH_UNKNOWN]: '!border-l-cool-grey-600 dark:!border-l-cool-grey-500',
  [NO_SIGNAL]: '!border-l-cool-grey-600 dark:!border-l-cool-grey-500',
}

export const healthFilterLabel = (value: string) =>
  value === NO_SIGNAL ? 'No health signal' : humanize(value)

export interface IHealthFilterDropdown {
  selected: Set<string>
  onChange: (selected: Set<string>) => void
}

export const HealthFilterOptions = ({
  selected,
  onChange,
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
    <>
      {HEALTH_FILTER_OPTIONS.map((value) => {
        const checked = selected.has(value)
        const isOnlySelected = selected.size === 1 && checked
        return (
          <CheckboxInputWithButton
            key={value}
            className={cn(
              'w-full border-l-4 pl-2 pr-0',
              HEALTH_BORDER_CLASSES[value]
            )}
            name={`fleet-health-${value}`}
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
                    {healthFilterLabel(value)}
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
    </>
  )
}

export const HealthFilterDropdown = ({
  selected,
  onChange,
}: IHealthFilterDropdown) => (
  <Dropdown
    buttonText={
      <span className="flex items-center gap-2">
        Health
        {selected.size ? (
          <Badge size="sm" theme="info">
            {selected.size}
          </Badge>
        ) : null}
      </span>
    }
    icon={<Icon variant="FunnelIcon" size="14" />}
    iconAlignment="left"
    id="fleet-filter-health"
    alignment="right"
    closeOnBlur={false}
  >
    <Menu className="min-w-56">
      <HealthFilterOptions selected={selected} onChange={onChange} />
      {selected.size ? <hr /> : null}
      {selected.size ? (
        <Button variant="ghost" onClick={() => onChange(new Set())}>
          Reset
        </Button>
      ) : null}
    </Menu>
  </Dropdown>
)
