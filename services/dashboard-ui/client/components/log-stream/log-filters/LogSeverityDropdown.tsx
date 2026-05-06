'use client'

import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'
import { cn } from '@/utils/classnames'
import { getSeverityBorderClasses } from '@/utils/log-stream-utils'

const LOG_SEVERITIES: {
  label: 'Trace' | 'Debug' | 'Info' | 'Warn' | 'Error' | 'Fatal'
  value: number
}[] = [
  { label: 'Trace', value: 4 },
  { label: 'Debug', value: 8 },
  { label: 'Info', value: 12 },
  { label: 'Warn', value: 16 },
  { label: 'Error', value: 20 },
  { label: 'Fatal', value: 24 },
]

interface ILogSeverityDropdown {
  filters: TLogFiltersProps
}

export const LogSeverityDropdown = ({ filters }: ILogSeverityDropdown) => {
  const {
    selectedSeverities,
    handleSeverityInputToggle,
    handleSeverityButtonClick,
    handleSeverityReset,
    severityStats,
  } = filters

  const showCount = !severityStats.isDefault && selectedSeverities.size > 0

  return (
    <Dropdown
      buttonText={
        <span className="flex items-center gap-2">
          Severity
          {showCount ? (
            <Badge size="sm" theme="info">
              {selectedSeverities.size}
            </Badge>
          ) : null}
        </span>
      }
      icon={<Icon variant="FunnelIcon" size="14" />}
      iconAlignment="left"
      id="log-severity-filter"
      alignment="right"
      closeOnBlur={false}
    >
      <Menu className="!w-56">
        {LOG_SEVERITIES.map((s) => {
          const checked = selectedSeverities.has(s.label)
          const isOnlySelected =
            selectedSeverities.size === 1 && checked
          return (
            <CheckboxInputWithButton
              key={s.label}
              className={cn(
                'w-full border-l-4 pl-2 pr-2',
                getSeverityBorderClasses(s.value, 'l')
              )}
              name={s.label}
              value={s.label}
              checked={checked}
              onChange={() => handleSeverityInputToggle(s.label)}
              buttonProps={{
                className:
                  '!py-1 !pl-0 !pr-1 flex items-center justify-between group w-full',
                variant: 'ghost',
                type: 'button',
                value: s.label,
                onClick: () => handleSeverityButtonClick(s.label),
                children: (
                  <>
                    <Text variant="body" weight="strong">
                      {s.label}
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

        {showCount && <hr />}
        {showCount && (
          <Button variant="ghost" onClick={handleSeverityReset}>
            Reset
          </Button>
        )}
      </Menu>
    </Dropdown>
  )
}
