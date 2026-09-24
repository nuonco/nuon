import { createContext, useMemo, useState } from 'react'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useStoredViewMode } from '@/hooks/use-stored-view-mode'
import {
  getInstallsTabEnabled,
  getStatusBarEnabled,
  setInstallsTabEnabled,
  setStatusBarEnabled,
} from '@/lib/cookies'

export type TDiffViewer = 'legacy' | 'v2'
export type TDiffWrap = 'scroll' | 'wrap'
export type TDiffView = 'unified' | 'split'
export type TPlanSections = 'collapsed' | 'expanded'

export const DIFF_VIEWERS: readonly TDiffViewer[] = ['legacy', 'v2']
export const DIFF_WRAPS: readonly TDiffWrap[] = ['scroll', 'wrap']
export const DIFF_VIEWS: readonly TDiffView[] = ['unified', 'split']
export const PLAN_SECTIONS: readonly TPlanSections[] = [
  'collapsed',
  'expanded',
]

export const SHOW_IDS_STORAGE_KEY = 'nuon-show-ids'

export const DEFAULT_DIFF_VIEWER: TDiffViewer = 'legacy'
export const DEFAULT_DIFF_WRAP: TDiffWrap = 'scroll'
export const DEFAULT_DIFF_VIEW: TDiffView = 'unified'
export const DEFAULT_PLAN_SECTIONS: TPlanSections = 'collapsed'

function readStoredShowIds(): boolean | undefined {
  try {
    const stored = localStorage.getItem(SHOW_IDS_STORAGE_KEY)
    if (stored === 'shown') return true
    if (stored === 'hidden') return false
  } catch {}
  return undefined
}

function writeStoredShowIds(showIds: boolean | undefined) {
  try {
    if (showIds === undefined) localStorage.removeItem(SHOW_IDS_STORAGE_KEY)
    else localStorage.setItem(SHOW_IDS_STORAGE_KEY, showIds ? 'shown' : 'hidden')
  } catch {}
}

export interface IDashboardPreferencesContext {
  showIds: boolean
  setShowIds: (showIds: boolean) => void
  isStatusBarEnabled: boolean
  setIsStatusBarEnabled: (isEnabled: boolean) => void
  isInstallsTabEnabled: boolean
  setIsInstallsTabEnabled: (isEnabled: boolean) => void
  diffViewer: TDiffViewer
  setDiffViewer: (viewer: TDiffViewer) => void
  diffWrap: TDiffWrap
  setDiffWrap: (wrap: TDiffWrap) => void
  diffView: TDiffView
  setDiffView: (view: TDiffView) => void
  planSections: TPlanSections
  setPlanSections: (sections: TPlanSections) => void
  resetPreferences: () => void
}

export const DashboardPreferencesContext = createContext<
  IDashboardPreferencesContext | undefined
>(undefined)

export const DashboardPreferencesProvider = ({
  children,
}: {
  children: React.ReactNode
}) => {
  const { installsTabAutoEnabled, statusBarAutoEnabled } = useConfig()
  const { isNuonEmployee } = useAuth()
  const [storedShowIds, setStoredShowIds] = useState(readStoredShowIds)
  const showIds = storedShowIds ?? isNuonEmployee
  const [isStatusBarEnabled, setIsStatusBarEnabledState] = useState(
    () => getStatusBarEnabled() ?? statusBarAutoEnabled ?? true
  )
  const [isInstallsTabEnabled, setIsInstallsTabEnabledState] = useState(
    () => getInstallsTabEnabled() ?? installsTabAutoEnabled ?? true
  )
  const [diffViewer, setDiffViewer] = useStoredViewMode<TDiffViewer>(
    'nuon-diff-viewer',
    DIFF_VIEWERS,
    DEFAULT_DIFF_VIEWER
  )
  const [diffWrap, setDiffWrap] = useStoredViewMode<TDiffWrap>(
    'nuon-diff-wrap',
    DIFF_WRAPS,
    DEFAULT_DIFF_WRAP
  )
  const [diffView, setDiffView] = useStoredViewMode<TDiffView>(
    'nuon-diff-view',
    DIFF_VIEWS,
    DEFAULT_DIFF_VIEW
  )
  const [planSections, setPlanSections] = useStoredViewMode<TPlanSections>(
    'nuon-plan-sections',
    PLAN_SECTIONS,
    DEFAULT_PLAN_SECTIONS
  )

  const value = useMemo(
    () => ({
      showIds,
      setShowIds: (next: boolean) => {
        writeStoredShowIds(next)
        setStoredShowIds(next)
      },
      isStatusBarEnabled,
      setIsStatusBarEnabled: (isEnabled: boolean) => {
        setStatusBarEnabled(isEnabled)
        setIsStatusBarEnabledState(isEnabled)
      },
      isInstallsTabEnabled,
      setIsInstallsTabEnabled: (isEnabled: boolean) => {
        setInstallsTabEnabled(isEnabled)
        setIsInstallsTabEnabledState(isEnabled)
      },
      diffViewer,
      setDiffViewer,
      diffWrap,
      setDiffWrap,
      diffView,
      setDiffView,
      planSections,
      setPlanSections,
      resetPreferences: () => {
        writeStoredShowIds(undefined)
        setStoredShowIds(undefined)
        setStatusBarEnabled(true)
        setIsStatusBarEnabledState(true)
        setInstallsTabEnabled(true)
        setIsInstallsTabEnabledState(true)
        setDiffViewer(DEFAULT_DIFF_VIEWER)
        setDiffWrap(DEFAULT_DIFF_WRAP)
        setDiffView(DEFAULT_DIFF_VIEW)
        setPlanSections(DEFAULT_PLAN_SECTIONS)
      },
    }),
    [
      showIds,
      isInstallsTabEnabled,
      isStatusBarEnabled,
      diffViewer,
      setDiffViewer,
      diffWrap,
      setDiffWrap,
      diffView,
      setDiffView,
      planSections,
      setPlanSections,
    ]
  )

  return (
    <DashboardPreferencesContext.Provider value={value}>
      {children}
    </DashboardPreferencesContext.Provider>
  )
}
