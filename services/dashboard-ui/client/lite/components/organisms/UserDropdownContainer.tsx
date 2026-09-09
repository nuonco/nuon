import { useSurfaces } from '../../hooks/use-surfaces'
import { UserDropdown, type IUserDropdown } from './UserDropdown'
import { UserPreferencesPanel } from './UserPreferencesPanel'

export const UserDropdownContainer = (props: IUserDropdown) => {
  const { openPanel } = useSurfaces()

  return (
    <UserDropdown
      {...props}
      onOpenPreferences={() => openPanel(<UserPreferencesPanel />)}
    />
  )
}
