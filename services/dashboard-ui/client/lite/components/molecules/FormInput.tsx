import type { AnyFieldApi } from '@tanstack/react-form'
import { Input, type IInput } from '../atoms/Input'
import { Field, type IField } from './Field'
import { fieldErrorMessage } from './field-error'

export interface IFormInput
  extends Omit<
      IInput,
      'value' | 'defaultValue' | 'onChange' | 'onBlur' | 'name'
    >,
    Pick<IField, 'label' | 'description' | 'optional' | 'className'> {
  field: AnyFieldApi
}

export const FormInput = ({
  field,
  label,
  description,
  optional,
  className,
  ...props
}: IFormInput) => (
  <Field
    label={label}
    description={description}
    optional={optional}
    error={fieldErrorMessage(field)}
    className={className}
  >
    <Input
      name={field.name}
      value={(field.state.value as string | undefined) ?? ''}
      onChange={(event) => field.handleChange(event.target.value)}
      onBlur={field.handleBlur}
      {...props}
    />
  </Field>
)
