import { useCallback, useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { BranchSettingsPanelContent } from './BranchSettingsPanelContent'

export const BRANCH_SETTINGS_PANEL_KEY = 'branch-settings'

export const useOpenBranchSettings = () => {
  const [, setSearchParams] = useSearchParams()
  return useCallback(() => {
    const params = new URLSearchParams(window.location.search)
    params.set('panel', BRANCH_SETTINGS_PANEL_KEY)
    setSearchParams(params, { replace: true })
  }, [setSearchParams])
}

export const BranchSettingsPanel = () => {
  const { addPanel } = useSurfaces()
  const [searchParams] = useSearchParams()
  const panelParam = searchParams.get('panel')
  const openIdRef = useRef<string | null>(null)

  useEffect(() => {
    if (panelParam !== BRANCH_SETTINGS_PANEL_KEY) {
      openIdRef.current = null
      return
    }
    if (openIdRef.current) return
    const timer = setTimeout(() => {
      openIdRef.current = addPanel(
        <Panel
          heading={
            <span className="flex flex-col">
              <Text flex weight="strong" variant="h3">
                <Icon variant="GearIcon" /> Settings
              </Text>
              <Text variant="subtext" theme="neutral">
                Manage this branch&apos;s name, source, triggers, previews, and
                lifecycle.
              </Text>
            </span>
          }
          size="half"
        >
          <BranchSettingsPanelContent />
        </Panel>,
        BRANCH_SETTINGS_PANEL_KEY
      )
    }, 0)
    return () => clearTimeout(timer)
  }, [panelParam, addPanel])

  return null
}
