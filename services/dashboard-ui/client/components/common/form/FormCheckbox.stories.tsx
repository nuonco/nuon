import { useForm } from '@tanstack/react-form'
import { FormCheckbox } from './FormCheckbox'

export default { title: 'Common/Forms/FormCheckbox' }

const Demo = () => {
  const form = useForm({ defaultValues: { autoApprove: false } })
  return (
    <div className="max-w-md p-4">
      <form.Field name="autoApprove">
        {(field) => (
          <FormCheckbox
            field={field}
            labelProps={{ labelText: 'Auto-approve changes' }}
          />
        )}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />
