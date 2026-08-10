import { useForm } from '@tanstack/react-form'
import { FormInput } from './FormInput'

export default { title: 'Common/Forms/FormInput' }

const Demo = () => {
  const form = useForm({ defaultValues: { name: '' } })
  return (
    <div className="max-w-md p-4">
      <form.Field name="name">
        {(field) => (
          <FormInput
            field={field}
            labelProps={{ labelText: 'Name' }}
            placeholder="e.g. ci-deploy"
          />
        )}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />
