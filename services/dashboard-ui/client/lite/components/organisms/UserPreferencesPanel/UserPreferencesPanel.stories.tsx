import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { useUserPreferences } from '../../../providers/user-preferences-provider'
import { UserPreferencesPanel } from './UserPreferencesPanel'

export default {
  title: 'lite/organisms/UserPreferencesPanel',
}

export const Overview = () => (
  <ComponentDocs
    name="UserPreferencesPanel"
    tier="organism"
    summary="Browser-wide defaults for Lite appearance, collections, and plan diffs."
    use={[
      'Open from the user dropdown when someone wants to change persistent interface defaults.',
      'Apply each choice immediately and keep the current page visible behind the panel.',
    ]}
    avoid={[
      'Do not use for organization or resource settings.',
      'Do not add a save step or route state.',
    ]}
    rules={[
      'Preferences apply across accounts and organizations in the same browser.',
      'Reset restores system theme, table view, scrolling lines, collapsed sections, and unified diffs.',
      'Plan expansion controls only the initial state of a newly mounted plan.',
    ]}
    props={[
      {
        name: 'preferences',
        type: 'IUserPreferences',
        description: 'Current validated preference values.',
      },
      {
        name: 'onPreferenceChange',
        type: '(key, value) => void',
        description: 'Applies one preference immediately.',
      },
      {
        name: 'onReset',
        type: '() => void',
        description: 'Restores every preference default.',
      },
    ]}
  />
)

const PreferencesDemo = () => {
  const { preferences, setPreference, resetPreferences } = useUserPreferences()

  return (
    <SurfaceStory
      open={({ openPanel }) =>
        openPanel(
          <UserPreferencesPanel
            preferences={preferences}
            onPreferenceChange={setPreference}
            onReset={resetPreferences}
          />
        )
      }
    />
  )
}

export const Default = () => <PreferencesDemo />
