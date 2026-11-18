export default {
  title: 'Features / Users / Preferences panel',
}

import { useState } from 'react'
import { UserPreferencesPanel } from './UserPreferencesPanel'
import type {
  TDiffView,
  TDiffViewer,
  TDiffWrap,
  TPlanSections,
  TTableVersion,
} from '@/providers/dashboard-preferences-provider'
import type { TThemePreference } from '@/providers/theme-provider'

const Demo = ({
  initialViewer = 'legacy',
  initialTableVersion = 'legacy',
}: {
  initialViewer?: TDiffViewer
  initialTableVersion?: TTableVersion
}) => {
  const [theme, setTheme] = useState<TThemePreference>('system')
  const [showIds, setShowIds] = useState(false)
  const [installsTab, setInstallsTab] = useState(true)
  const [statusBar, setStatusBar] = useState(true)
  const [tableVersion, setTableVersion] =
    useState<TTableVersion>(initialTableVersion)
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
      showIds={showIds}
      onShowIdsChange={setShowIds}
      isInstallsTabEnabled={installsTab}
      onInstallsTabChange={setInstallsTab}
      isStatusBarEnabled={statusBar}
      onStatusBarChange={setStatusBar}
      tableVersion={tableVersion}
      onTableVersionChange={setTableVersion}
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
        setShowIds(false)
        setInstallsTab(true)
        setStatusBar(true)
        setTableVersion('legacy')
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

export const NewTablesSelected = () => <Demo initialTableVersion="v2" />
