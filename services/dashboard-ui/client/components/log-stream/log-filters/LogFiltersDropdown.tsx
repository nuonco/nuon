import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { RadioInput } from '@/components/common/form/RadioInput'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

const LABEL_PROPS = { variant: 'body', weight: 'strong' } as const

interface ILogFiltersDropdown {
  filters: TLogFiltersProps
}

export const LogFiltersDropdown = ({ filters }: ILogFiltersDropdown) => {
  const {
    includeSystemLogs,
    handleSystemLogsToggle,
    sortStats,
    handleSortToggle,
    tool,
    setTool,
    availableTools,
    isFiltered,
    handleResetAll,
  } = filters

  const activeCount = (includeSystemLogs ? 1 : 0) + (tool ? 1 : 0)

  const toolOptions = Array.from(
    tool && !availableTools.has(tool)
      ? new Set([...availableTools, tool])
      : availableTools
  ).sort()

  return (
    <Dropdown
      buttonText={
        <span className="flex items-center gap-2">
          Filter
          {activeCount > 0 && (
            <Badge size="sm" theme="info">
              {activeCount}
            </Badge>
          )}
        </span>
      }
      icon={<Icon variant="FunnelIcon" size="14" />}
      iconAlignment="left"
      id="log-filters-combined"
      alignment="right"
      closeOnBlur={false}
    >
      <Menu className="!w-72">
        <SectionHeading>Sort</SectionHeading>
        <RadioInput
          name="log-sort"
          checked={sortStats.isNewestFirst}
          onChange={handleSortToggle}
          labelProps={{
            labelText: 'Latest first',
            labelTextProps: LABEL_PROPS,
          }}
        />
        <RadioInput
          name="log-sort"
          checked={sortStats.isOldestFirst}
          onChange={handleSortToggle}
          labelProps={{
            labelText: 'Oldest first',
            labelTextProps: LABEL_PROPS,
          }}
        />

        {toolOptions.length > 0 && (
          <>
            <hr />
            <SectionHeading>Tool</SectionHeading>
            <RadioInput
              name="log-tool"
              checked={!tool}
              onChange={() => setTool('')}
              labelProps={{ labelText: 'All', labelTextProps: LABEL_PROPS }}
            />
            {toolOptions.map((t) => (
              <RadioInput
                key={t}
                name="log-tool"
                checked={tool === t}
                onChange={() => setTool(tool === t ? '' : t)}
                labelProps={{ labelText: t, labelTextProps: LABEL_PROPS }}
              />
            ))}
          </>
        )}

        <hr />

        <SectionHeading>Debug</SectionHeading>
        <CheckboxInput
          checked={includeSystemLogs}
          onChange={handleSystemLogsToggle}
          labelProps={{
            labelText: 'Include system logs',
            labelTextProps: LABEL_PROPS,
          }}
        />

        {isFiltered && <hr />}
        {isFiltered && (
          <Button variant="ghost" onClick={handleResetAll}>
            Reset
          </Button>
        )}
      </Menu>
    </Dropdown>
  )
}

const SectionHeading = ({ children }: { children: ReactNode }) => (
  <div className="px-2 pt-1 pb-0.5">
    <Text variant="label" weight="strong" theme="neutral">
      {children}
    </Text>
  </div>
)
