'use client'

import type React from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import {
  CheckboxInput,
  CheckboxInputWithButton,
} from '@/components/common/form/CheckboxInput'
import { RadioInput } from '@/components/common/form/RadioInput'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

const LOG_SERVICES = ['api', 'runner'] as const

// Per Stratus design system "Body Strong": body size + medium weight
const BODY_STRONG = { variant: 'body', weight: 'strong' } as const

interface ILogFiltersDropdown {
  filters: TLogFiltersProps
}

export const LogFiltersDropdown = ({ filters }: ILogFiltersDropdown) => {
  const {
    jobOutputOnly,
    handleJobOutputToggle,
    sortStats,
    handleSortToggle,
    tool,
    setTool,
    availableTools,
    selectedScopes,
    availableScopes,
    handleScopeInputToggle,
    handleScopeButtonClick,
    selectedServices,
    handleServiceInputToggle,
    handleServiceButtonClick,
    isFiltered,
    handleResetAll,
  } = filters

  const activeCount = computeActiveFilterCount(filters)

  const toolOptions = Array.from(
    new Set<string>([...availableTools, ...(tool ? [tool] : [])])
  ).sort()
  const scopeOptions = Array.from(
    new Set<string>([...availableScopes, ...selectedScopes])
  ).sort()

  return (
    <Dropdown
      buttonText={
        <span className="flex items-center gap-2">
          Filter
          {activeCount > 0 ? (
            <Badge size="sm" theme="info">
              {activeCount}
            </Badge>
          ) : null}
        </span>
      }
      icon={<Icon variant="FunnelIcon" size="14" />}
      iconAlignment="left"
      id="log-filters-combined"
      alignment="right"
      closeOnBlur={false}
    >
      <Menu className="!w-72 max-h-[28rem] overflow-y-auto">
        <SectionHeading>Display</SectionHeading>
        <CheckboxInput
          checked={jobOutputOnly}
          onChange={handleJobOutputToggle}
          labelProps={{
            labelText: 'Job output only',
            labelTextProps: BODY_STRONG,
          }}
        />

        <hr />

        <SectionHeading>Sort</SectionHeading>
        <RadioInput
          name="log-sort"
          checked={sortStats.isNewestFirst}
          onChange={() => {
            if (!sortStats.isNewestFirst) handleSortToggle()
          }}
          labelProps={{
            labelText: 'Latest first',
            labelTextProps: BODY_STRONG,
          }}
        />
        <RadioInput
          name="log-sort"
          checked={sortStats.isOldestFirst}
          onChange={() => {
            if (!sortStats.isOldestFirst) handleSortToggle()
          }}
          labelProps={{
            labelText: 'Oldest first',
            labelTextProps: BODY_STRONG,
          }}
        />

        {toolOptions.length > 0 ? (
          <>
            <hr />
            <SectionHeading>Tool</SectionHeading>
            <RadioInput
              name="log-tool"
              checked={!tool}
              onChange={() => setTool('')}
              labelProps={{ labelText: 'All', labelTextProps: BODY_STRONG }}
            />
            {toolOptions.map((t) => (
              <RadioInput
                key={t}
                name="log-tool"
                checked={tool === t}
                onChange={() => setTool(tool === t ? '' : t)}
                labelProps={{ labelText: t, labelTextProps: BODY_STRONG }}
              />
            ))}
          </>
        ) : null}

        {scopeOptions.length > 0 ? (
          <>
            <hr />
            <SectionHeading>Scope</SectionHeading>
            {scopeOptions.map((s) => {
              const checked = selectedScopes.has(s)
              const isOnlySelected = selectedScopes.size === 1 && checked
              return (
                <CheckboxInputWithButton
                  key={s}
                  className="w-full px-2"
                  name={s}
                  value={s}
                  checked={checked}
                  onChange={() => handleScopeInputToggle(s)}
                  buttonProps={{
                    className:
                      '!py-1 !pl-0 !pr-1 flex items-center justify-between group w-full',
                    variant: 'ghost',
                    type: 'button',
                    value: s,
                    onClick: () => handleScopeButtonClick(s),
                    children: (
                      <>
                        <Text variant="body" weight="strong">
                          {s}
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
        ) : null}

        <hr />

        <SectionHeading>Service</SectionHeading>
        {LOG_SERVICES.map((s) => {
          const checked = selectedServices.has(s)
          const isOnlySelected = selectedServices.size === 1 && checked
          return (
            <CheckboxInputWithButton
              key={s}
              className="w-full px-2"
              name={s}
              value={s}
              checked={checked}
              onChange={() => handleServiceInputToggle(s)}
              buttonProps={{
                className:
                  '!py-1 !pl-0 !pr-1 flex items-center justify-between group w-full',
                variant: 'ghost',
                type: 'button',
                value: s,
                onClick: () => handleServiceButtonClick(s),
                children: (
                  <>
                    <Text variant="body" weight="strong">
                      {s}
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

function computeActiveFilterCount(filters: TLogFiltersProps): number {
  let count = 0
  if (filters.jobOutputOnly) count += 1
  if (filters.tool) count += 1
  count += filters.selectedScopes.size
  count += filters.selectedServices.size
  return count
}

const SectionHeading = ({ children }: { children: React.ReactNode }) => (
  <div className="px-2 pt-1 pb-0.5">
    <Text variant="label" weight="strong" theme="neutral">
      {children}
    </Text>
  </div>
)
