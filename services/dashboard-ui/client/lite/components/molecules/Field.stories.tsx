import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Input } from '../atoms/Input'
import { Field } from './Field'

export default { title: 'lite/molecules/Field' }

export const Overview = () => (
  <ComponentDocs
    name="Field"
    tier="molecule"
    summary="One shared layout for a label, description, control and validation error."
    use={[
      'Wrap Input, Textarea and Select.',
      'Let Form wrappers supply its validation error.',
    ]}
    avoid={[
      'Do not recreate label or error spacing inside individual controls.',
    ]}
    rules={[
      'Field owns ids and accessible descriptions.',
      'Optional is descriptive; validation still belongs to Zod.',
    ]}
    props={[
      {
        name: 'label',
        type: 'ReactNode',
        description: 'Visible control label.',
      },
      {
        name: 'description',
        type: 'ReactNode',
        description: 'Guidance before the control.',
      },
      {
        name: 'error',
        type: 'ReactNode',
        description: 'Validation error after the control.',
      },
      {
        name: 'optional',
        type: 'boolean',
        default: 'false',
        description: 'Shows the optional hint.',
      },
    ]}
  />
)

export const States = () => (
  <div className="grid max-w-3xl gap-6 p-8 sm:grid-cols-2">
    <Field label="Install name">
      <Input placeholder="acme-production" />
    </Field>
    <Field label="Description" optional>
      <Input placeholder="What this install is for" />
    </Field>
    <Field
      label="Region"
      description="The cloud region where resources are deployed."
    >
      <Input defaultValue="us-west-2" />
    </Field>
    <Field
      label="Install name"
      error="Use lowercase letters, numbers and hyphens"
    >
      <Input defaultValue="Production Install" />
    </Field>
  </div>
)

export const WithoutLabel = () => (
  <div className="max-w-sm p-8">
    <Field description="A label can live elsewhere when another element already names the control.">
      <Input aria-label="Filter installs" placeholder="Filter installs" />
    </Field>
  </div>
)
