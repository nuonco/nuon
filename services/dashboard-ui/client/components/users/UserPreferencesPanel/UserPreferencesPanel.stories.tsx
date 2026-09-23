export default {
  title: 'Users/UserPreferencesPanel',
}

import { useState } from 'react'
import { UserPreferencesPanel } from './UserPreferencesPanel'
import type {
  TDiffView,
  TDiffViewer,
  TDiffWrap,
  TPlanSections,
} from '@/providers/dashboard-preferences-provider'
import type { TThemePreference } from '@/providers/theme-provider'

const Demo = ({
  initialViewer = 'legacy',
}: {
  initialViewer?: TDiffViewer
}) => {
  const [theme, setTheme] = useState<TThemePreference>('system')
  const [installsTab, setInstallsTab] = useState(true)
  const [statusBar, setStatusBar] = useState(true)
  const [diffViewer, setDiffViewer] = useState<TDiffViewer>(initialViewer)
  const [diffView, setDiffView] = useState<TDiffView>('unified')
  const [diffWrap, setDiffWrap] = useState<TDiffWrap>('scroll')
  const [planSections, setPlanSections] = useState<TPlanSections>('collapsed')

  return (
    <UserPreferencesPanel
      isVisible
      panelId="preferences-story"
      theme={theme}
      onThemeChange={setTheme}
      isInstallsTabEnabled={installsTab}
      onInstallsTabChange={setInstallsTab}
      isStatusBarEnabled={statusBar}
      onStatusBarChange={setStatusBar}
      diffViewer={diffViewer}
      onDiffViewerChange={setDiffViewer}
      diffView={diffView}
      onDiffViewChange={setDiffView}
      diffWrap={diffWrap}
      onDiffWrapChange={setDiffWrap}
      planSections={planSections}
      onPlanSectionsChange={setPlanSections}
      onResetPreferences={() => {
        setTheme('system')
        setInstallsTab(true)
        setStatusBar(true)
        setDiffViewer('legacy')
        setDiffView('unified')
        setDiffWrap('scroll')
        setPlanSections('collapsed')
      }}
    />
  )
}

export const Default = () => <Demo />

export const NewDiffViewerSelected = () => <Demo initialViewer="v2" />
