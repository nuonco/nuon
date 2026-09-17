import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Field } from '../molecules/Field'
import { Input } from './Input'
import { Text } from './Text'

export default { title: 'lite/atoms/Input' }

export const Overview = () => (
  <ComponentDocs
    name="Input"
    tier="atom"
    summary="A native single-line input with lite sizing, focus, invalid and loading states."
    use={[
      'Use for short free text, numbers, URLs, email and passwords.',
      'Wrap in Field for visible label and validation.',
    ]}
    avoid={[
      'Do not pass native required.',
      'Do not put helper or error text inside Input.',
    ]}
    rules={[
      'Pass native input types through.',
      'Use md in forms and sm in compact filters.',
    ]}
    props={[
      {
        name: 'size',
        type: "'sm' | 'md'",
        default: "'md'",
        description: 'Control height and text size.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders the same control box as a loading state.',
      },
    ]}
  />
)

export const Sizes = () => (
  <div className="flex max-w-md flex-col gap-4 p-8">
    <Field label="Small">
      <Input size="sm" placeholder="Filter resources" />
    </Field>
    <Field label="Medium">
      <Input placeholder="Install name" />
    </Field>
  </div>
)

export const Types = () => (
  <div className="grid max-w-3xl gap-4 p-8 sm:grid-cols-2">
    <Field label="Email">
      <Input type="email" placeholder="operator@example.com" />
    </Field>
    <Field label="Password">
      <Input type="password" defaultValue="not-a-secret" />
    </Field>
    <Field label="Port">
      <Input type="number" defaultValue={8080} />
    </Field>
    <Field label="Endpoint">
      <Input type="url" placeholder="https://api.example.com" />
    </Field>
  </div>
)

export const States = () => (
  <div className="grid max-w-3xl gap-4 p-8 sm:grid-cols-2">
    <Field label="Default">
      <Input placeholder="Type a value" />
    </Field>
    <Field label="Filled">
      <Input defaultValue="payments-api" />
    </Field>
    <Field label="Disabled">
      <Input disabled defaultValue="Managed by app config" />
    </Field>
    <Field label="Loading" loading>
      <Input loading />
    </Field>
    <Field label="Invalid" error="Enter a valid hostname">
      <Input defaultValue="not a hostname" />
    </Field>
    <Field label="Read only">
      <Input readOnly value="org_01h9k2m4p6" />
    </Field>
  </div>
)

export const LongValue = () => (
  <div className="max-w-xs p-8">
    <Field label="Image">
      <Input defaultValue="registry.example.com/platform/payments-api:2026.09.02-release-candidate" />
    </Field>
    <Text variant="caption" color="tertiary" className="mt-3">
      The native input scrolls horizontally without widening its container.
    </Text>
  </div>
)
