import type { ChangeEvent } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import type { TTriggerAuthType, TTriggerEnvelope } from '@/types'

export const AUTH_TYPE_OPTIONS: TTriggerAuthType[] = [
  'none',
  'hmac',
  'api_key',
  'basic',
  'bearer_jwt',
  'sns_signature',
]

export const ENVELOPE_OPTIONS: TTriggerEnvelope[] = [
  'none',
  'pubsub_push',
  'cloudevents',
  'sns',
]

export const SOURCE_OPTIONS = [
  'GitHub',
  'Slack',
  'Datadog',
  'AWS',
  'GCP',
  'Azure',
  'Custom',
]

const FilterDropdown = ({
  id,
  label,
  name,
  options,
  value,
  onChange,
  onClear,
}: {
  id: string
  label: string
  name: string
  options: string[]
  value: string
  onChange: (e: ChangeEvent<HTMLInputElement>) => void
  onClear: () => void
}) => (
  <Dropdown
    alignment="right"
    buttonText={
      <>
        <Icon variant="FunnelIcon" size="14" />
        {value ? `${label}: ${value}` : label}
      </>
    }
    id={id}
  >
    <Menu className="!p-0 !w-56">
      <form onReset={onClear}>
        <div className="flex flex-col gap-0.5 max-h-[250px] overflow-y-auto w-full p-2 focus-visible:outline-1 focus-visible:outline-primary-600 rounded-md">
          {options.map((option) => (
            <RadioInput
              key={option}
              checked={value === option}
              labelProps={{
                labelText: option,
                labelTextProps: { family: 'mono' },
              }}
              name={name}
              onChange={onChange}
              value={option}
            />
          ))}
        </div>

        <div className="flex flex-col gap-0.5 px-2 pb-2 w-full">
          <hr />
          <Button className="mt-1" isMenuButton type="reset" variant="ghost">
            Clear
            <Icon variant="XIcon" />
          </Button>
        </div>
      </form>
    </Menu>
  </Dropdown>
)

export interface ITriggerFilters {
  trigger: string
  authType: string
  envelope: string
  onSourceChange: (e: ChangeEvent<HTMLInputElement>) => void
  onAuthTypeChange: (e: ChangeEvent<HTMLInputElement>) => void
  onEnvelopeChange: (e: ChangeEvent<HTMLInputElement>) => void
  onClearSource: () => void
  onClearAuthType: () => void
  onClearEnvelope: () => void
}

export const TriggerFilters = ({
  trigger,
  authType,
  envelope,
  onSourceChange,
  onAuthTypeChange,
  onEnvelopeChange,
  onClearSource,
  onClearAuthType,
  onClearEnvelope,
}: ITriggerFilters) => (
  <div className="flex items-center gap-4 flex-wrap">
    <FilterDropdown
      id="event-trigger-trigger-filter"
      label="Source"
      name="event-trigger-trigger"
      options={SOURCE_OPTIONS}
      value={trigger}
      onChange={onSourceChange}
      onClear={onClearSource}
    />
    <FilterDropdown
      id="event-trigger-auth-type-filter"
      label="Auth type"
      name="event-trigger-auth-type"
      options={AUTH_TYPE_OPTIONS}
      value={authType}
      onChange={onAuthTypeChange}
      onClear={onClearAuthType}
    />
    <FilterDropdown
      id="event-trigger-envelope-filter"
      label="Envelope"
      name="event-trigger-envelope"
      options={ENVELOPE_OPTIONS}
      value={envelope}
      onChange={onEnvelopeChange}
      onClear={onClearEnvelope}
    />
  </div>
)
