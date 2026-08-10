import { useForm } from '@tanstack/react-form'
import { FormCodeInput } from './FormCodeInput'

export default { title: 'Common/Forms/FormCodeInput' }

const Demo = () => {
  const form = useForm({ defaultValues: { config: '' } })
  return (
    <div className="max-w-md p-4">
      <form.Field name="config">
        {(field) => (
          <FormCodeInput
            field={field}
            language="yaml"
            labelProps={{ labelText: 'Config' }}
          />
        )}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />
