import { useForm } from '@tanstack/react-form'
import { FormToggle } from './FormToggle'

export default { title: 'Common/Forms/FormToggle' }

const Demo = () => {
  const form = useForm({ defaultValues: { enabled: false } })
  return (
    <div className="max-w-md p-4">
      <form.Field name="enabled">
        {(field) => (
          <FormToggle
            field={field}
            label="Enabled"
            description="Turn this policy on or off"
          />
        )}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />
