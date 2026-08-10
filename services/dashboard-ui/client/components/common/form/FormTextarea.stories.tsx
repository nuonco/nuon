import { useForm } from '@tanstack/react-form'
import { FormTextarea } from './FormTextarea'

export default { title: 'Common/Forms/FormTextarea' }

const Demo = () => {
  const form = useForm({ defaultValues: { notes: '' } })
  return (
    <div className="max-w-md p-4">
      <form.Field name="notes">
        {(field) => (
          <FormTextarea
            field={field}
            labelProps={{ labelText: 'Notes' }}
            placeholder="Add a description…"
          />
        )}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />
