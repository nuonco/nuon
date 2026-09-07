import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Radio } from './Radio'
import { Text } from './Text'

export default { title: 'lite/atoms/Radio' }

export const Overview = () => (
  <ComponentDocs
    name="Radio"
    tier="atom"
    summary="One option in a mutually exclusive set."
    use={[
      'Use in a named group when exactly one option may be selected.',
      'Use FormRadioGroup for form-bound sets.',
    ]}
    avoid={[
      'Do not render a single radio.',
      'Do not use for independent choices. Use Checkbox.',
    ]}
    rules={[
      'Every option has a visible label.',
      'Descriptions explain the consequence of that option.',
    ]}
    props={[
      {
        name: 'label',
        type: 'ReactNode',
        description: 'Clickable option label.',
      },
      {
        name: 'description',
        type: 'ReactNode',
        description: 'Supporting text below the label.',
      },
      { name: 'error', type: 'ReactNode', description: 'Validation error.' },
    ]}
  />
)

const GroupDemo = () => {
  const [value, setValue] = useState('manual')
  return (
    <fieldset className="flex flex-col gap-3">
      <Text as="legend" variant="label" color="secondary" className="mb-2">
        Approval mode
      </Text>
      <Radio
        name="approval-mode"
        value="manual"
        checked={value === 'manual'}
        onChange={() => setValue('manual')}
        label="Manual"
        description="Wait for an operator to approve every plan."
      />
      <Radio
        name="approval-mode"
        value="automatic"
        checked={value === 'automatic'}
        onChange={() => setValue('automatic')}
        label="Automatic"
        description="Deploy matching plans as soon as they are ready."
      />
      <Radio
        name="approval-mode"
        value="policy"
        disabled
        label="Policy based"
        description="Available after an approval policy is configured."
      />
    </fieldset>
  )
}

export const Group = () => (
  <div className="max-w-md p-8">
    <GroupDemo />
  </div>
)

export const States = () => (
  <div className="flex flex-col gap-5 p-8">
    <Radio name="states" label="Unchecked" />
    <Radio name="states-checked" label="Checked" defaultChecked />
    <Radio name="states-disabled" label="Disabled" disabled />
    <Radio
      name="states-disabled-checked"
      label="Disabled and checked"
      disabled
      defaultChecked
    />
    <Radio name="states-loading" label="Loading" loading />
    <Radio
      name="states-error"
      label="Invalid option"
      error="Choose another option"
    />
  </div>
)
