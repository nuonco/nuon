import { useUserPreferences } from '../../../providers/user-preferences-provider'
import { UserPreferencesPanel } from './UserPreferencesPanel'

export const UserPreferencesPanelContainer = () => {
  const { preferences, setPreference, resetPreferences } = useUserPreferences()

  return (
    <UserPreferencesPanel
      preferences={preferences}
      onPreferenceChange={setPreference}
      onReset={resetPreferences}
    />
  )
}
