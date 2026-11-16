import { useForm } from '@tanstack/react-form'
import { FormRadioGroup } from './FormRadioGroup'

export default { title: 'UI / Forms / Form radio group' }

const OPTIONS = [
  { value: 'all', label: 'Everything in this org' },
  { value: 'specific', label: 'Specific resources' },
]

const OPTIONS_WITH_DESCRIPTIONS = [
  {
    value: 'create',
    label: 'Create a new principal',
    description: 'Setup creates a dedicated IAM role for this connection.',
  },
  {
    value: 'existing',
    label: 'Use an existing principal',
    description: 'Reuse a role another connection already uses.',
  },
]

const Demo = ({
  options,
  label,
}: {
  options: typeof OPTIONS
  label: string
}) => {
  const form = useForm({ defaultValues: { mode: 'all' } })
  return (
    <div className="max-w-md p-4">
      <form.Field name="mode">
        {(field) => (
          <FormRadioGroup field={field} label={label} options={options} />
        )}
      </form.Field>
    </div>
  )
}

export const Default = () => (
  <Demo options={OPTIONS} label="What should this match?" />
)
export const WithDescriptions = () => (
  <Demo options={OPTIONS_WITH_DESCRIPTIONS} label="Principal" />
)
