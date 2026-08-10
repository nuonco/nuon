import { useForm } from '@tanstack/react-form'
import { allEvents } from './defaults'
import { FormInterestsPicker } from './FormInterestsPicker'

export default { title: 'Interests/FormInterestsPicker' }

const Demo = () => {
  const form = useForm({ defaultValues: { interests: allEvents() } })
  return (
    <div className="max-w-lg p-4">
      <form.Field name="interests">
        {(field) => <FormInterestsPicker field={field} />}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />
