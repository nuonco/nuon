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

const preferencesSchema = z.object({
  theme: z.enum(['system', 'light', 'dark', 'classic', 'high-contrast']),
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
    installsTab: isInstallsTabEnabled ? ('shown' as const) : ('hidden' as const),
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
        <div className="flex flex-col gap-2">
          <form.Field
            name="theme"
            listeners={{
              onChange: ({ value }) =>
                onThemeChange(value as TThemePreference),
            }}
          >
            {(field) => (
              <FormRadioGroup
                field={field}
                label="Appearance"
                options={[
                  { value: 'system', label: 'System' },
                  { value: 'light', label: 'Light' },
                  { value: 'dark', label: 'Dark' },
                  { value: 'classic', label: 'Classic' },
                  { value: 'high-contrast', label: 'High contrast' },
                ]}
              />
            )}
          </form.Field>
          <Text variant="subtext" theme="neutral">
            How the dashboard looks in this browser. Classic is the previous
            purple theme.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <form.Field
            name="installsTab"
            listeners={{
              onChange: ({ value }) => onInstallsTabChange(value === 'shown'),
            }}
          >
            {(field) => (
              <FormRadioGroup
                field={field}
                label="Installs page"
                options={[
                  { value: 'shown', label: 'Shown' },
                  { value: 'hidden', label: 'Hidden' },
                ]}
              />
            )}
          </form.Field>
          <Text variant="subtext" theme="neutral">
            Show the org-level installs page in the main navigation.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <form.Field
            name="statusBar"
            listeners={{
              onChange: ({ value }) => onStatusBarChange(value === 'shown'),
            }}
          >
            {(field) => (
              <FormRadioGroup
                field={field}
                label="Status bar"
                options={[
                  { value: 'shown', label: 'Shown' },
                  { value: 'hidden', label: 'Hidden' },
                ]}
              />
            )}
          </form.Field>
          <Text variant="subtext" theme="neutral">
            Show the persistent status bar along the bottom of the dashboard.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <form.Field
            name="diffViewer"
            listeners={{
              onChange: ({ value }) =>
                onDiffViewerChange(value as TDiffViewer),
            }}
          >
            {(field) => (
              <FormRadioGroup
                field={field}
                label="Plan diff viewer"
                options={[
                  { value: 'legacy', label: 'Current' },
                  { value: 'v2', label: 'New' },
                ]}
              />
            )}
          </form.Field>
          <Text variant="subtext" theme="neutral">
            The new viewer adds search, chunk collapsing and split view. The
            plan graphs and attribute tree are only in the current viewer.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
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
                label="Plan diff view"
                options={[
                  { value: 'unified', label: 'Unified' },
                  { value: 'split', label: 'Split' },
                ]}
              />
            )}
          </form.Field>
          <Text variant="subtext" theme="neutral">
            Show before and after in one column, or side by side.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
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
                label="Plan diff lines"
                options={[
                  { value: 'scroll', label: 'Scroll' },
                  { value: 'wrap', label: 'Wrap' },
                ]}
              />
            )}
          </form.Field>
          <Text variant="subtext" theme="neutral">
            Keep long lines on one row, or wrap them to the available width.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
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
                label="Plan sections"
                options={[
                  { value: 'collapsed', label: 'Collapsed' },
                  { value: 'expanded', label: 'Expanded' },
                ]}
              />
            )}
          </form.Field>
          <Text variant="subtext" theme="neutral">
            Whether plan sections start open when a plan loads.
          </Text>
        </div>
      </form>
    </Panel>
  )
}
