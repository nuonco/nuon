import { ThemeSwitcher } from './ThemeSwitcher'
import { Text } from '../atoms/Text'
import { ComponentDocs } from '../__stories__/ComponentDocs'

export default {
  title: 'lite/molecules/ThemeSwitcher',
}

export const Overview = () => (
  <ComponentDocs
    name="ThemeSwitcher"
    tier="molecule"
    summary="Picks the theme preference: light, dark, high contrast, or system."
    use={[
      'Anywhere the user should be able to change the theme — settings, the account menu.',
    ]}
    avoid={['Reading the current theme for display. Use useTheme() for that.']}
    rules={[
      'It sets the preference, not the theme. system is a preference that resolves to light or dark and tracks the OS live.',
      'The preference is persisted in localStorage under nuon-lite-theme.',
      'Renders as a radiogroup, so arrow keys and screen readers work.',
    ]}
    sections={[
      {
        heading: 'In Ladle',
        body: "Ladle's own light/dark/auto control is wired to the same preference, so the toolbar toggle themes lite components for real. High contrast is only reachable through this component.",
      },
    ]}
  />
)

export const Default = () => <ThemeSwitcher />

export const OnSurface = () => (
  <div className="flex flex-col gap-4 rounded-xl border border-divider bg-surface-01 p-6">
    <Text variant="caption" color="tertiary">
      The switcher sits on a raised surface, so the selected pill reads against
      it rather than the page background.
    </Text>
    <ThemeSwitcher />
  </div>
)
