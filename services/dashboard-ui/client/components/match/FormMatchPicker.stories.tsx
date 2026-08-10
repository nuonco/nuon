import { useForm } from '@tanstack/react-form'
import { FormMatchPicker } from './FormMatchPicker'

export default { title: 'Match/FormMatchPicker' }

const Demo = () => {
  const form = useForm({ defaultValues: { match: undefined } })
  return (
    <div className="max-w-lg p-4">
      <form.Field name="match">
        {(field) => <FormMatchPicker field={field} />}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />
