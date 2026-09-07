import type { AnyFieldApi } from '@tanstack/react-form'
import { Field, type IField } from './Field'
import { Select, type ISelect } from './Select'
import { fieldErrorMessage } from './field-error'

export interface IFormSelect
  extends Omit<
      ISelect,
      'value' | 'defaultValue' | 'onChange' | 'onBlur' | 'name'
    >,
    Pick<IField, 'label' | 'description' | 'optional' | 'className'> {
  field: AnyFieldApi
}

export const FormSelect = ({
  field,
  label,
  description,
  optional,
  className,
  ...props
}: IFormSelect) => (
  <Field
    label={label}
    description={description}
    optional={optional}
    error={fieldErrorMessage(field)}
    className={className}
  >
    <Select
      name={field.name}
      value={(field.state.value as string | undefined) ?? ''}
      onChange={(value) => field.handleChange(value)}
      onBlur={field.handleBlur}
      {...props}
    />
  </Field>
)
