import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Checkbox } from './Checkbox'

export default { title: 'lite/atoms/Checkbox' }

export const Overview = () => (
  <ComponentDocs
    name="Checkbox"
    tier="atom"
    summary="A binary choice with a large label hit target and optional supporting text."
    use={[
      'Use when multiple independent choices may be selected.',
      'Use indeterminate for a partially selected parent set.',
    ]}
    avoid={[
      'Do not use for mutually exclusive options. Use Radio.',
      'Do not use for an immediate setting. Use Switch.',
    ]}
    rules={[
      'The label row is the hit target.',
      'Errors sit with the choice they describe.',
    ]}
    props={[
      {
        name: 'label',
        type: 'ReactNode',
        description: 'Clickable choice label.',
      },
      {
        name: 'description',
        type: 'ReactNode',
        description: 'Supporting text below the label.',
      },
      {
        name: 'error',
        type: 'ReactNode',
        description: 'Validation error below the description.',
      },
      {
        name: 'indeterminate',
        type: 'boolean',
        default: 'false',
        description: 'Shows partial selection.',
      },
    ]}
  />
)

const InteractiveDemo = () => {
  const [checked, setChecked] = useState(false)
  return (
    <Checkbox
      checked={checked}
      onChange={(event) => setChecked(event.target.checked)}
      label="Include job output"
      description="Download logs emitted by the running job."
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
    <Checkbox label="Unchecked" />
    <Checkbox label="Checked" defaultChecked />
    <Checkbox label="Partially selected" indeterminate />
    <Checkbox label="Disabled" disabled />
    <Checkbox label="Disabled and checked" disabled defaultChecked />
    <Checkbox label="Loading" loading />
    <Checkbox label="Terms accepted" error="Accept the terms to continue" />
  </div>
)

export const LongDescription = () => (
  <div className="max-w-md p-8">
    <Checkbox
      label="Automatically approve matching plans"
      description="Plans that only update image tags and match this policy will deploy without waiting for manual approval."
    />
  </div>
)

export const WithoutLabel = () => (
  <div className="p-8">
    <Checkbox aria-label="Select payments-api" />
  </div>
)
