import { useForm } from '@tanstack/react-form'
import { FormRadioGroup } from './FormRadioGroup'

export default { title: 'Common/Forms/FormRadioGroup' }

const OPTIONS = [
  { value: 'all', label: 'Everything in this org' },
  { value: 'specific', label: 'Specific resources' },
]

const Demo = () => {
  const form = useForm({ defaultValues: { mode: 'all' } })
  return (
    <div className="max-w-md p-4">
      <form.Field name="mode">
        {(field) => (
          <FormRadioGroup
            field={field}
            label="What should this match?"
            options={OPTIONS}
          />
        )}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />
