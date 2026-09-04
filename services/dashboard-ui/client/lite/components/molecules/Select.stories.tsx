import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Field } from './Field'
import { Select, type ISelectOption } from './Select'

export default { title: 'lite/molecules/Select' }

const REGIONS: ISelectOption[] = [
  {
    value: 'us-east-1',
    label: 'US East (N. Virginia)',
    description: 'us-east-1',
  },
  { value: 'us-east-2', label: 'US East (Ohio)', description: 'us-east-2' },
  {
    value: 'us-west-1',
    label: 'US West (N. California)',
    description: 'us-west-1',
  },
  { value: 'us-west-2', label: 'US West (Oregon)', description: 'us-west-2' },
  { value: 'eu-west-1', label: 'Europe (Ireland)', description: 'eu-west-1' },
  {
    value: 'eu-central-1',
    label: 'Europe (Frankfurt)',
    description: 'eu-central-1',
  },
  {
    value: 'ap-southeast-1',
    label: 'Asia Pacific (Singapore)',
    description: 'ap-southeast-1',
  },
  {
    value: 'ap-southeast-2',
    label: 'Asia Pacific (Sydney)',
    description: 'ap-southeast-2',
  },
]

export const Overview = () => (
  <ComponentDocs
    name="Select"
    tier="molecule"
    summary="A value-bearing listbox positioned by Dropdown."
    use={[
      'Use for choosing one value from a known option list.',
      'Enable search for longer lists.',
    ]}
    avoid={[
      'Do not use MenuItem for options.',
      'Do not use this for remote org, app or branch comboboxes.',
    ]}
    rules={[
      'Arrow keys, Home, End and typeahead move through options.',
      'Descriptions clarify options without becoming their values.',
    ]}
    props={[
      {
        name: 'options',
        type: 'ISelectOption[]',
        description: 'Known value, label, description and disabled state.',
      },
      {
        name: 'value',
        type: 'string',
        description: 'Controlled selected value.',
      },
      {
        name: 'defaultValue',
        type: 'string',
        description: 'Initial uncontrolled value.',
      },
      {
        name: 'onChange',
        type: '(value: string) => void',
        description: 'Runs after selection.',
      },
      {
        name: 'searchable',
        type: 'boolean',
        default: 'false',
        description: 'Adds local option filtering.',
      },
      {
        name: 'size',
        type: "'sm' | 'md'",
        default: "'md'",
        description: 'Trigger size.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders the trigger loading state.',
      },
    ]}
  />
)

const ControlledDemo = () => {
  const [value, setValue] = useState('us-west-2')
  return (
    <Field label="Region" description={`Selected value: ${value}`}>
      <Select value={value} onChange={setValue} options={REGIONS} />
    </Field>
  )
}

export const Default = () => (
  <div className="max-w-sm p-8">
    <ControlledDemo />
  </div>
)

export const Sizes = () => (
  <div className="flex max-w-sm flex-col gap-5 p-8">
    <Field label="Small">
      <Select size="sm" options={REGIONS} placeholder="Select region" />
    </Field>
    <Field label="Medium">
      <Select options={REGIONS} placeholder="Select region" />
    </Field>
  </div>
)

export const Searchable = () => (
  <div className="max-w-sm p-8">
    <Field label="Region" description="Type a region name or identifier.">
      <Select searchable options={REGIONS} placeholder="Select region" />
    </Field>
  </div>
)

export const DisabledOptions = () => (
  <div className="max-w-sm p-8">
    <Field label="Runner type">
      <Select
        options={[
          { value: 'aws', label: 'AWS', description: 'Available' },
          { value: 'azure', label: 'Azure', description: 'Available' },
          {
            value: 'gcp',
            label: 'Google Cloud',
            description: 'Coming later',
            disabled: true,
          },
        ]}
      />
    </Field>
  </div>
)

export const States = () => (
  <div className="grid max-w-3xl gap-5 p-8 sm:grid-cols-2">
    <Field label="Placeholder">
      <Select options={REGIONS} />
    </Field>
    <Field label="Selected">
      <Select options={REGIONS} defaultValue="eu-west-1" />
    </Field>
    <Field label="Disabled">
      <Select options={REGIONS} defaultValue="us-west-2" disabled />
    </Field>
    <Field label="Loading" loading>
      <Select options={REGIONS} loading />
    </Field>
    <Field label="Invalid" error="Choose a region">
      <Select options={REGIONS} />
    </Field>
    <Field label="No options">
      <Select options={[]} emptyMessage="No regions available" />
    </Field>
  </div>
)

export const LongLabels = () => (
  <div className="max-w-xs p-8">
    <Field label="Role">
      <Select
        defaultValue="long"
        options={[
          {
            value: 'long',
            label:
              'A very long role name that must truncate inside a narrow form',
            description:
              'Still keeps the trigger and menu inside their containers.',
          },
          { value: 'short', label: 'Operator' },
        ]}
      />
    </Field>
  </div>
)
