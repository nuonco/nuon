import type { ReactNode } from 'react'
import { Text } from '@/components/common/Text'
import { RadioInput } from '@/components/common/form/RadioInput'
import type { TAppFleetRunnerFilter } from '@/types'
import { FleetFilterDropdown } from './FleetFilterDropdown'

const LABEL_PROPS = { variant: 'body', weight: 'strong' } as const

const RUNNER_OPTIONS = [
  { value: 'active', label: 'Active' },
  { value: 'not-active', label: 'Not active' },
] as const

export interface IFleetFiltersDropdown {
  componentChoices?: { value: string; label: string }[]
  component?: string
  onComponentChange?: (value?: string) => void
  runner?: TAppFleetRunnerFilter
  onRunnerChange: (value?: TAppFleetRunnerFilter) => void
  onReset: () => void
}

export const FleetFiltersDropdown = ({
  componentChoices = [],
  component,
  onComponentChange = () => undefined,
  runner,
  onRunnerChange,
  onReset,
}: IFleetFiltersDropdown) => (
  <FleetFilterDropdown
    id="fleet-filters"
    label="Filters"
    count={(component ? 1 : 0) + (runner ? 1 : 0)}
    onReset={onReset}
    menuClassName="!w-64"
  >
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
    {componentChoices.length ? (
      <>
        <hr className="my-1" />
        <SectionHeading>Compare runs on</SectionHeading>
        <RadioChoices
          name="fleet-component"
          allLabel="Whole install"
          options={componentChoices}
          selected={component}
          onChange={onComponentChange}
        />
      </>
    ) : null}
  </FleetFilterDropdown>
)

const SectionHeading = ({ children }: { children: ReactNode }) => (
  <div className="px-2 pt-1 pb-0.5">
    <Text variant="label" weight="strong" theme="neutral">
      {children}
    </Text>
  </div>
)

const RadioChoices = ({
  name,
  allLabel = 'All',
  options,
  selected,
  onChange,
}: {
  name: string
  allLabel?: string
  options: readonly { value: string; label: string }[]
  selected?: string
  onChange: (value?: string) => void
}) => (
  <>
    <RadioInput
      name={name}
      checked={!selected}
      onChange={() => onChange(undefined)}
      labelProps={{ labelText: allLabel, labelTextProps: LABEL_PROPS }}
    />
    {options.map((option) => (
      <RadioInput
        key={option.value}
        name={name}
        checked={selected === option.value}
        onChange={() => onChange(option.value)}
        labelProps={{ labelText: option.label, labelTextProps: LABEL_PROPS }}
      />
    ))}
  </>
)
