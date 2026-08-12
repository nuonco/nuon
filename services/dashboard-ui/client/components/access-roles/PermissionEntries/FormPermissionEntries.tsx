import type { AnyFieldApi } from '@tanstack/react-form'
import { PermissionEntriesContainer } from './PermissionEntriesContainer'

export const FormPermissionEntries = ({
  field,
  disabled,
}: {
  field: AnyFieldApi
  disabled?: boolean
}) => (
  <PermissionEntriesContainer
    value={field.state.value ?? []}
    onChange={(next) => field.handleChange(next)}
    disabled={disabled}
  />
)
