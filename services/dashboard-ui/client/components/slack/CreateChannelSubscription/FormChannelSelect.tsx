import type { AnyFieldApi } from '@tanstack/react-form'
import { ChannelSelect, type IChannelSelect } from './ChannelSelect'

export interface IFormChannelSelect
  extends Omit<IChannelSelect, 'value' | 'onChange'> {
  field: AnyFieldApi
  onName?: (name: string) => void
}

export const FormChannelSelect = ({
  field,
  onName,
  ...props
}: IFormChannelSelect) => (
  <ChannelSelect
    value={(field.state.value as string | undefined) ?? ''}
    onChange={(id, name) => {
      field.handleChange(id)
      onName?.(name)
    }}
    {...props}
  />
)
