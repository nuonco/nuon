import { useEffect } from 'react'
import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import {
  DEFAULT_USER_PREFERENCES,
  type IUserPreferences,
  type TCollectionView,
  type TDiffPreferenceView,
  type TThemePreference,
} from '../../../providers/user-preferences-provider'
import { Button } from '../../atoms/Button'
import { FormRadioGroup } from '../../molecules/FormRadioGroup'
import { Panel } from '../surfaces'

const preferencesSchema = z.object({
  theme: z.enum(['light', 'dark', 'high-contrast', 'system']),
  collectionView: z.enum(['table', 'cards']),
  diffWrap: z.enum(['scroll', 'wrap']),
  planSections: z.enum(['collapsed', 'expanded']),
  diffView: z.enum(['unified', 'split']),
})

const formValues = (preferences: IUserPreferences) => ({
  theme: preferences.theme,
  collectionView: preferences.collectionView,
  diffWrap: preferences.diffWrap ? ('wrap' as const) : ('scroll' as const),
  planSections: preferences.planSectionsOpen
    ? ('expanded' as const)
    : ('collapsed' as const),
  diffView: preferences.diffView,
})

export interface IUserPreferencesPanel {
  preferences: IUserPreferences
  onPreferenceChange: <K extends keyof IUserPreferences>(
    key: K,
    value: IUserPreferences[K]
  ) => void
  onReset: () => void
}

export const UserPreferencesPanel = ({
  preferences,
  onPreferenceChange,
  onReset,
}: IUserPreferencesPanel) => {
  const form = useForm({
    defaultValues: formValues(preferences),
    validators: {
      onMount: preferencesSchema,
      onChange: preferencesSchema,
    },
  })

  useEffect(() => {
    form.reset(formValues(preferences))
  }, [form, preferences])

  const reset = () => {
    form.reset(formValues(DEFAULT_USER_PREFERENCES))
    onReset()
  }

  return (
    <Panel
      heading="Preferences"
      expandable={false}
      headerActions={
        <Button size="sm" variant="ghost" onClick={reset}>
          Reset preferences
        </Button>
      }
    >
      <form
        autoComplete="off"
        noValidate
        className="flex flex-col gap-8"
        onSubmit={(event) => event.preventDefault()}
      >
        <form.Field name="theme">
          {(field) => (
            <FormRadioGroup
              field={field}
              label="Appearance"
              description="Choose how Lite looks in this browser."
              options={[
                {
                  value: 'system',
                  label: 'System',
                  description: 'Match your operating system appearance.',
                },
                {
                  value: 'light',
                  label: 'Light',
                  description: 'Always use the light theme.',
                },
                {
                  value: 'dark',
                  label: 'Dark',
                  description: 'Always use the dark theme.',
                },
                {
                  value: 'high-contrast',
                  label: 'High contrast',
                  description: 'Use stronger contrast for interface elements.',
                },
              ]}
              onValueChange={(value) =>
                onPreferenceChange('theme', value as TThemePreference)
              }
            />
          )}
        </form.Field>

        <form.Field name="collectionView">
          {(field) => (
            <FormRadioGroup
              field={field}
              label="Collections"
              description="Set the default presentation for lists that support both views."
              options={[
                {
                  value: 'table',
                  label: 'Table',
                  description: 'Show rows and columns when space allows.',
                },
                {
                  value: 'cards',
                  label: 'Cards',
                  description:
                    'Show every supported collection as a card grid.',
                },
              ]}
              onValueChange={(value) =>
                onPreferenceChange('collectionView', value as TCollectionView)
              }
            />
          )}
        </form.Field>

        <form.Field name="diffWrap">
          {(field) => (
            <FormRadioGroup
              field={field}
              label="Plan diff lines"
              description="Choose how long lines appear in plan diffs."
              options={[
                {
                  value: 'scroll',
                  label: 'Scroll',
                  description:
                    'Keep each line on one row and scroll horizontally.',
                },
                {
                  value: 'wrap',
                  label: 'Wrap',
                  description: 'Wrap long lines to fit the available width.',
                },
              ]}
              onValueChange={(value) =>
                onPreferenceChange('diffWrap', value === 'wrap')
              }
            />
          )}
        </form.Field>

        <form.Field name="planSections">
          {(field) => (
            <FormRadioGroup
              field={field}
              label="Plan sections"
              description="Set the initial state when a plan opens."
              options={[
                {
                  value: 'collapsed',
                  label: 'Collapsed',
                  description: 'Start plan sections closed.',
                },
                {
                  value: 'expanded',
                  label: 'Expanded',
                  description: 'Start plan sections open.',
                },
              ]}
              onValueChange={(value) =>
                onPreferenceChange('planSectionsOpen', value === 'expanded')
              }
            />
          )}
        </form.Field>

        <form.Field name="diffView">
          {(field) => (
            <FormRadioGroup
              field={field}
              label="Plan diff view"
              description="Choose how before and after content is arranged."
              options={[
                {
                  value: 'unified',
                  label: 'Unified',
                  description: 'Show additions and removals in one column.',
                },
                {
                  value: 'split',
                  label: 'Split',
                  description: 'Show before and after content side by side.',
                },
              ]}
              onValueChange={(value) =>
                onPreferenceChange('diffView', value as TDiffPreferenceView)
              }
            />
          )}
        </form.Field>
      </form>
    </Panel>
  )
}
