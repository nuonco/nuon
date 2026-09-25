import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { Button } from '@/components/common/Button'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { Text } from '@/components/common/Text'
import type {
  TDiffView,
  TDiffViewer,
  TDiffWrap,
  TPlanSections,
} from '@/providers/dashboard-preferences-provider'
import type { TThemePreference } from '@/providers/theme-provider'

const THEME_OPTIONS: Array<{
  value: TThemePreference
  label: string
  swatch: string
}> = [
  {
    value: 'system',
    label: 'System',
    swatch: 'linear-gradient(145deg, #f4f4f5 0%, #8a8a90 46%, #141416 100%)',
  },
  {
    value: 'light',
    label: 'Light',
    swatch: 'linear-gradient(145deg, #ffffff 0%, #d9f6fd 52%, #4cc9f0 100%)',
  },
  {
    value: 'dark',
    label: 'Dark',
    swatch: 'linear-gradient(145deg, #101012 0%, #1a4554 52%, #4cc9f0 100%)',
  },
  {
    value: 'classic',
    label: 'Classic',
    swatch: 'linear-gradient(145deg, #1a1220 0%, #5b2d84 52%, #c084fc 100%)',
  },
  {
    value: 'high-contrast',
    label: 'High contrast',
    swatch: 'linear-gradient(145deg, #0000aa 0%, #2f2fe0 42%, #ffff66 100%)',
  },
  {
    value: 'monochrome',
    label: 'Monochrome',
    swatch: 'linear-gradient(145deg, #0a0a0a 0%, #f5f5f5 58%, #4cc9f0 100%)',
  },
]

const sectionLabel = (label: string) => (
  <Text variant="label" theme="neutral">
    {label}
  </Text>
)

const sectionDescription = (description: string) => (
  <Text variant="subtext" theme="neutral">
    {description}
  </Text>
)

const preferencesSchema = z.object({
  theme: z.enum([
    'system',
    'light',
    'dark',
    'classic',
    'high-contrast',
    'monochrome',
  ]),
  showIds: z.enum(['hidden', 'shown']),
  installsTab: z.enum(['hidden', 'shown']),
  statusBar: z.enum(['hidden', 'shown']),
  diffViewer: z.enum(['legacy', 'v2']),
  diffView: z.enum(['unified', 'split']),
  diffWrap: z.enum(['scroll', 'wrap']),
  planSections: z.enum(['collapsed', 'expanded']),
})

export interface IUserPreferencesPanel extends IPanel {
  theme: TThemePreference
  onThemeChange: (theme: TThemePreference) => void
  showIds: boolean
  onShowIdsChange: (showIds: boolean) => void
  isInstallsTabEnabled: boolean
  onInstallsTabChange: (isEnabled: boolean) => void
  isStatusBarEnabled: boolean
  onStatusBarChange: (isEnabled: boolean) => void
  diffViewer: TDiffViewer
  onDiffViewerChange: (viewer: TDiffViewer) => void
  diffView: TDiffView
  onDiffViewChange: (view: TDiffView) => void
  diffWrap: TDiffWrap
  onDiffWrapChange: (wrap: TDiffWrap) => void
  planSections: TPlanSections
  onPlanSectionsChange: (sections: TPlanSections) => void
  onResetPreferences: () => void
}

export const UserPreferencesPanel = ({
  theme,
  onThemeChange,
  showIds,
  onShowIdsChange,
  isInstallsTabEnabled,
  onInstallsTabChange,
  isStatusBarEnabled,
  onStatusBarChange,
  diffViewer,
  onDiffViewerChange,
  diffView,
  onDiffViewChange,
  diffWrap,
  onDiffWrapChange,
  planSections,
  onPlanSectionsChange,
  onResetPreferences,
  ...props
}: IUserPreferencesPanel) => {
  const values = {
    theme,
    showIds: showIds ? ('shown' as const) : ('hidden' as const),
    installsTab: isInstallsTabEnabled
      ? ('shown' as const)
      : ('hidden' as const),
    statusBar: isStatusBarEnabled ? ('shown' as const) : ('hidden' as const),
    diffViewer,
    diffView,
    diffWrap,
    planSections,
  }

  const form = useForm({
    defaultValues: values,
    validators: {
      onMount: preferencesSchema,
      onChange: preferencesSchema,
    },
  })

  return (
    <Panel
      heading="Preferences"
      footer={
        <Button variant="secondary" onClick={onResetPreferences}>
          Reset preferences
        </Button>
      }
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        className="flex flex-col gap-8"
        onSubmit={(event) => event.preventDefault()}
      >
        <form.Field
          name="theme"
          listeners={{
            onChange: ({ value }) => onThemeChange(value as TThemePreference),
          }}
        >
          {(field) => (
            <FormRadioGroup
              field={field}
              label={sectionLabel('Appearance')}
              description={sectionDescription(
                'How the dashboard looks in this browser. Classic is the previous purple theme. Monochrome is black and white, with color kept for status.'
              )}
              options={THEME_OPTIONS.map((option) => ({
                value: option.value,
                label: (
                  <span className="inline-flex items-center gap-2">
                    <span
                      aria-hidden
                      className="inline-block size-5 shrink-0 rounded-full"
                      style={{
                        background: option.swatch,
                        boxShadow:
                          'inset 0 0 0 1px rgba(255, 255, 255, 0.22), 0 0 0 1px rgba(0, 0, 0, 0.35)',
                      }}
                    />
                    {option.label}
                  </span>
                ),
              }))}
            />
          )}
        </form.Field>

        <form.Field
          name="showIds"
          listeners={{
            onChange: ({ value }) => onShowIdsChange(value === 'shown'),
          }}
        >
          {(field) => (
            <FormRadioGroup
              field={field}
              label={sectionLabel('Resource IDs')}
              description={sectionDescription(
                'Show the raw resource ID under names on cards, tables and headers.'
              )}
              options={[
                { value: 'shown', label: 'Shown' },
                { value: 'hidden', label: 'Hidden' },
              ]}
            />
          )}
        </form.Field>

        <form.Field
          name="installsTab"
          listeners={{
            onChange: ({ value }) => onInstallsTabChange(value === 'shown'),
          }}
        >
          {(field) => (
            <FormRadioGroup
              field={field}
              label={sectionLabel('Installs page')}
              description={sectionDescription(
                'Show the org-level installs page in the main navigation.'
              )}
              options={[
                { value: 'shown', label: 'Shown' },
                { value: 'hidden', label: 'Hidden' },
              ]}
            />
          )}
        </form.Field>

        <form.Field
          name="statusBar"
          listeners={{
            onChange: ({ value }) => onStatusBarChange(value === 'shown'),
          }}
        >
          {(field) => (
            <FormRadioGroup
              field={field}
              label={sectionLabel('Status bar')}
              description={sectionDescription(
                'Show the persistent status bar along the bottom of the dashboard.'
              )}
              options={[
                { value: 'shown', label: 'Shown' },
                { value: 'hidden', label: 'Hidden' },
              ]}
            />
          )}
        </form.Field>

        <form.Field
          name="diffViewer"
          listeners={{
            onChange: ({ value }) => onDiffViewerChange(value as TDiffViewer),
          }}
        >
          {(field) => (
            <FormRadioGroup
              field={field}
              label={sectionLabel('Plan diff viewer')}
              description={sectionDescription(
                'The new viewer adds search, chunk collapsing and split view. The plan graphs and attribute tree are only in the current viewer.'
              )}
              options={[
                { value: 'legacy', label: 'Current' },
                { value: 'v2', label: 'New' },
              ]}
            />
          )}
        </form.Field>

        <form.Field
          name="diffView"
          listeners={{
            onChange: ({ value }) => onDiffViewChange(value as TDiffView),
          }}
        >
          {(field) => (
            <FormRadioGroup
              field={field}
              disabled={diffViewer === 'legacy'}
              label={sectionLabel('Plan diff view')}
              description={sectionDescription(
                'Show before and after in one column, or side by side.'
              )}
              options={[
                { value: 'unified', label: 'Unified' },
                { value: 'split', label: 'Split' },
              ]}
            />
          )}
        </form.Field>

        <form.Field
          name="diffWrap"
          listeners={{
            onChange: ({ value }) => onDiffWrapChange(value as TDiffWrap),
          }}
        >
          {(field) => (
            <FormRadioGroup
              field={field}
              disabled={diffViewer === 'legacy'}
              label={sectionLabel('Plan diff lines')}
              description={sectionDescription(
                'Keep long lines on one row, or wrap them to the available width.'
              )}
              options={[
                { value: 'scroll', label: 'Scroll' },
                { value: 'wrap', label: 'Wrap' },
              ]}
            />
          )}
        </form.Field>

        <form.Field
          name="planSections"
          listeners={{
            onChange: ({ value }) =>
              onPlanSectionsChange(value as TPlanSections),
          }}
        >
          {(field) => (
            <FormRadioGroup
              field={field}
              disabled={diffViewer === 'legacy'}
              label={sectionLabel('Plan sections')}
              description={sectionDescription(
                'Whether plan sections start open when a plan loads.'
              )}
              options={[
                { value: 'collapsed', label: 'Collapsed' },
                { value: 'expanded', label: 'Expanded' },
              ]}
            />
          )}
        </form.Field>
      </form>
    </Panel>
  )
}
