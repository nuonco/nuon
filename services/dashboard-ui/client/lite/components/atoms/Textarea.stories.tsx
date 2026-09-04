import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Field } from '../molecules/Field'
import { Textarea } from './Textarea'

export default { title: 'lite/atoms/Textarea' }

export const Overview = () => (
  <ComponentDocs
    name="Textarea"
    tier="atom"
    summary="A multiline text control with the same field surface as Input."
    use={[
      'Use for descriptions, notes and multiline plain text.',
      'Enable autoResize when the surrounding layout should grow with the value.',
    ]}
    avoid={[
      'Do not use for structured code. CodeInput remains separate.',
      'Do not pass native required.',
    ]}
    rules={[
      'Wrap in Field for labels and validation.',
      'Manual resize is vertical only.',
    ]}
    props={[
      {
        name: 'size',
        type: "'sm' | 'md'",
        default: "'md'",
        description: 'Padding and text size.',
      },
      {
        name: 'autoResize',
        type: 'boolean',
        default: 'false',
        description: 'Grows between minRows and maxRows.',
      },
      {
        name: 'minRows',
        type: 'number',
        default: '3',
        description: 'Initial and minimum row count.',
      },
      {
        name: 'maxRows',
        type: 'number',
        default: '10',
        description: 'Maximum auto-resize row count.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders the textarea-shaped loading state.',
      },
    ]}
  />
)

export const Sizes = () => (
  <div className="grid max-w-3xl gap-6 p-8 sm:grid-cols-2">
    <Field label="Small notes">
      <Textarea size="sm" placeholder="Add a note" />
    </Field>
    <Field label="Description">
      <Textarea placeholder="Describe this install" />
    </Field>
  </div>
)

export const States = () => (
  <div className="grid max-w-3xl gap-6 p-8 sm:grid-cols-2">
    <Field label="Default">
      <Textarea placeholder="Add context" />
    </Field>
    <Field label="Filled">
      <Textarea defaultValue="Deploys the payments service into the production account." />
    </Field>
    <Field label="Disabled">
      <Textarea disabled defaultValue="Managed by app config." />
    </Field>
    <Field label="Loading" loading>
      <Textarea loading />
    </Field>
    <Field label="Invalid" error="Description must be under 120 characters">
      <Textarea defaultValue="A description that needs editing." />
    </Field>
  </div>
)

const AutoResizeDemo = () => {
  const [value, setValue] = useState(
    'Start typing here.\nAdd lines to grow the control.'
  )
  return (
    <div className="max-w-md p-8">
      <Field
        label="Release notes"
        description="Grows from 2 to 6 rows, then scrolls."
      >
        <Textarea
          autoResize
          minRows={2}
          maxRows={6}
          value={value}
          onChange={(event) => setValue(event.target.value)}
        />
      </Field>
    </div>
  )
}

export const AutoResize = () => <AutoResizeDemo />
