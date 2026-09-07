import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Switch } from './Switch'

export default { title: 'lite/atoms/Switch' }

export const Overview = () => (
  <ComponentDocs
    name="Switch"
    tier="atom"
    summary="An immediate on/off setting, exposed with the switch role."
    use={[
      'Use for settings that take effect immediately.',
      'Use a label and description when the consequence is not obvious.',
    ]}
    avoid={[
      'Do not call it Toggle.',
      'Do not use for selecting items or agreeing to terms. Use Checkbox.',
    ]}
    rules={[
      'Checked state is controlled.',
      'The entire label row activates the switch.',
    ]}
    props={[
      {
        name: 'checked',
        type: 'boolean',
        description: 'Current switch state.',
      },
      {
        name: 'onChange',
        type: '(checked: boolean) => void',
        description: 'Receives the next state.',
      },
      { name: 'label', type: 'ReactNode', description: 'Setting label.' },
      {
        name: 'description',
        type: 'ReactNode',
        description: 'Consequence or supporting text.',
      },
      { name: 'error', type: 'ReactNode', description: 'Validation error.' },
    ]}
  />
)

const InteractiveDemo = () => {
  const [checked, setChecked] = useState(true)
  return (
    <Switch
      checked={checked}
      onChange={setChecked}
      label="Fallback polling"
      description="Poll every 30 seconds when the live stream disconnects."
    />
  )
}

export const Interactive = () => (
  <div className="p-8">
    <InteractiveDemo />
  </div>
)

export const States = () => (
  <div className="flex flex-col gap-5 p-8">
    <Switch checked={false} onChange={() => {}} label="Off" />
    <Switch checked onChange={() => {}} label="On" />
    <Switch checked={false} onChange={() => {}} label="Disabled" disabled />
    <Switch checked onChange={() => {}} label="Disabled and on" disabled />
    <Switch checked={false} onChange={() => {}} label="Loading" loading />
    <Switch
      checked={false}
      onChange={() => {}}
      label="Invalid setting"
      error="This setting is required by policy"
    />
  </div>
)

export const WithoutLabel = () => (
  <div className="p-8">
    <Switch checked onChange={() => {}} aria-label="Enable fallback polling" />
  </div>
)
