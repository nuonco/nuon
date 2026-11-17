import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { RadioInput } from '@/components/common/form/RadioInput'
import type { TComponentState, TFleetRunnerFilter } from './fleet'

const LABEL_PROPS = { variant: 'body', weight: 'strong' } as const

const ROLLOUT_OPTIONS = [
  { value: 'applied', label: 'Applied' },
  { value: 'behind', label: 'Not applied' },
] as const

const RUNNER_OPTIONS = [
  { value: 'active', label: 'Active' },
  { value: 'not-active', label: 'Not active' },
] as const

const LIVE_OPTIONS = [
  { value: 'live', label: 'Live' },
  { value: 'not-live', label: 'Not live' },
] as const

export interface IFleetFiltersDropdown {
  rollout?: 'applied' | 'behind'
  onRolloutChange: (value?: 'applied' | 'behind') => void
  componentChoices: { value: string; label: string }[]
  component?: string
  onComponentChange: (value?: string) => void
  componentState: TComponentState
  onComponentStateChange: (value: TComponentState) => void
  runner?: TFleetRunnerFilter
  onRunnerChange: (value?: TFleetRunnerFilter) => void
  onReset: () => void
}

export const FleetFiltersDropdown = ({
  rollout,
  onRolloutChange,
  componentChoices,
  component,
  onComponentChange,
  componentState,
  onComponentStateChange,
  runner,
  onRunnerChange,
  onReset,
}: IFleetFiltersDropdown) => {
  const activeCount = (rollout ? 1 : 0) + (component ? 1 : 0) + (runner ? 1 : 0)

  const sections = [
    <div key="rollout">
      <SectionHeading>Latest run</SectionHeading>
      <RadioChoices
        name="fleet-rollout"
        options={ROLLOUT_OPTIONS}
        selected={rollout}
        onChange={(value) =>
          onRolloutChange(
            value === 'applied' || value === 'behind' ? value : undefined
          )
        }
      />
    </div>,
    componentChoices.length ? (
      <div key="component">
        <SectionHeading>Component</SectionHeading>
        <RadioChoices
          name="fleet-component"
          options={componentChoices}
          selected={component}
          onChange={(value) => onComponentChange(value)}
        />
        {component ? (
          <>
            <SectionHeading>Live state</SectionHeading>
            {LIVE_OPTIONS.map((option) => (
              <RadioInput
                key={option.value}
                name="fleet-component-state"
                checked={componentState === option.value}
                onChange={() =>
                  onComponentStateChange(
                    option.value === 'not-live' ? 'not-live' : 'live'
                  )
                }
                labelProps={{
                  labelText: option.label,
                  labelTextProps: LABEL_PROPS,
                }}
              />
            ))}
          </>
        ) : null}
      </div>
    ) : null,
    <div key="runner">
      <SectionHeading>Runner</SectionHeading>
      <RadioChoices
        name="fleet-runner"
        options={RUNNER_OPTIONS}
        selected={runner}
        onChange={(value) =>
          onRunnerChange(
            value === 'active' || value === 'not-active' ? value : undefined
          )
        }
      />
    </div>,
  ].filter((section) => section !== null)

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
      id="fleet-filters"
      alignment="right"
      closeOnBlur={false}
    >
      <Menu className="!w-72 max-h-96">
        <div className="-mx-2 -mt-2 flex min-h-0 flex-col overflow-y-auto overflow-x-hidden px-2 pt-2">
          {sections.map((section, index) => (
            <div key={section.key}>
              {index > 0 ? <hr className="my-1" /> : null}
              {section}
            </div>
          ))}
        </div>
        <hr />
        <Button
          className="shrink-0"
          disabled={activeCount === 0}
          onClick={onReset}
          variant="ghost"
        >
          Reset
        </Button>
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

const RadioChoices = ({
  name,
  options,
  selected,
  onChange,
}: {
  name: string
  options: readonly { value: string; label: string }[]
  selected?: string
  onChange: (value?: string) => void
}) => (
  <>
    <RadioInput
      name={name}
      checked={!selected}
      onChange={() => onChange(undefined)}
      labelProps={{ labelText: 'All', labelTextProps: LABEL_PROPS }}
    />
    {options.map((option) => (
      <RadioInput
        key={option.value}
        name={name}
        checked={selected === option.value}
        onChange={() =>
          onChange(selected === option.value ? undefined : option.value)
        }
        labelProps={{ labelText: option.label, labelTextProps: LABEL_PROPS }}
      />
    ))}
  </>
)
