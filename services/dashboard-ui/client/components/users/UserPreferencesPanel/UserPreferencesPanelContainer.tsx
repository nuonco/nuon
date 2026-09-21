import { useDashboardPreferences } from '@/hooks/use-dashboard-preferences'
import { useTheme } from '@/hooks/use-theme'
import type { IPanel } from '@/components/surfaces/Panel'
import { UserPreferencesPanel } from './UserPreferencesPanel'

export const UserPreferencesPanelContainer = (props: IPanel) => {
  const { preference, setPreference } = useTheme()
  const {
    isInstallsTabEnabled,
    setIsInstallsTabEnabled,
    isStatusBarEnabled,
    setIsStatusBarEnabled,
    diffViewer,
    setDiffViewer,
    diffView,
    setDiffView,
    diffWrap,
    setDiffWrap,
    planSections,
    setPlanSections,
    resetPreferences,
  } = useDashboardPreferences()

  return (
    <UserPreferencesPanel
      theme={preference}
      onThemeChange={setPreference}
      isInstallsTabEnabled={isInstallsTabEnabled}
      onInstallsTabChange={setIsInstallsTabEnabled}
      isStatusBarEnabled={isStatusBarEnabled}
      onStatusBarChange={setIsStatusBarEnabled}
      diffViewer={diffViewer}
      onDiffViewerChange={setDiffViewer}
      diffView={diffView}
      onDiffViewChange={setDiffView}
      diffWrap={diffWrap}
      onDiffWrapChange={setDiffWrap}
      planSections={planSections}
      onPlanSectionsChange={setPlanSections}
      onResetPreferences={() => {
        setPreference('system')
        resetPreferences()
      }}
      {...props}
    />
  )
}
