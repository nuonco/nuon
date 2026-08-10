import { useForm } from '@tanstack/react-form'
import { FormSelect } from './FormSelect'

export default { title: 'Common/Forms/FormSelect' }

const OPTIONS = [
  { value: 'org_read_only', label: 'Read-only' },
  { value: 'org_builder', label: 'Builder' },
  { value: 'org_admin', label: 'Admin' },
]

const Demo = () => {
  const form = useForm({ defaultValues: { role: 'org_read_only' } })
  return (
    <div className="max-w-md p-4">
      <form.Field name="role">
        {(field) => (
          <FormSelect
            field={field}
            options={OPTIONS}
            labelProps={{ labelText: 'Role' }}
          />
        )}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />
