import type { AnyFieldApi } from '@tanstack/react-form'
import { Textarea, type ITextarea } from '../atoms/Textarea'
import { Field, type IField } from './Field'
import { fieldErrorMessage } from './field-error'

export interface IFormTextarea
  extends Omit<
      ITextarea,
      'value' | 'defaultValue' | 'onChange' | 'onBlur' | 'name'
    >,
    Pick<IField, 'label' | 'description' | 'optional' | 'className'> {
  field: AnyFieldApi
}

export const FormTextarea = ({
  field,
  label,
  description,
  optional,
  className,
  ...props
}: IFormTextarea) => (
  <Field
    label={label}
    description={description}
    optional={optional}
    error={fieldErrorMessage(field)}
    className={className}
  >
    <Textarea
      name={field.name}
      value={(field.state.value as string | undefined) ?? ''}
      onChange={(event) => field.handleChange(event.target.value)}
      onBlur={field.handleBlur}
      {...props}
    />
  </Field>
)
