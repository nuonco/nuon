import { useCallback, useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { InstallSettingsPanelContent } from './InstallSettingsPanelContent'

export const INSTALL_SETTINGS_PANEL_KEY = 'settings'

export const useOpenInstallSettings = () => {
  const [, setSearchParams] = useSearchParams()
  return useCallback(() => {
    const params = new URLSearchParams(window.location.search)
    params.set('panel', INSTALL_SETTINGS_PANEL_KEY)
    setSearchParams(params, { replace: true })
  }, [setSearchParams])
}

export const InstallSettingsPanel = () => {
  const { addPanel } = useSurfaces()
  const [searchParams] = useSearchParams()
  const panelParam = searchParams.get('panel')
  const openIdRef = useRef<string | null>(null)

  useEffect(() => {
    if (panelParam !== INSTALL_SETTINGS_PANEL_KEY) {
      openIdRef.current = null
      return
    }
    if (openIdRef.current) return
    // Defer so this runs after the pathname-change panel clear in SurfacesProvider
    const timer = setTimeout(() => {
      openIdRef.current = addPanel(
        <Panel
          heading={
            <span className="flex flex-col">
              <Text flex weight="strong" variant="h3">
                <Icon variant="GearIcon" /> Settings
              </Text>
              <Text variant="subtext" theme="neutral">
                Manage this install's configuration and lifecycle.
              </Text>
            </span>
          }
          size="half"
        >
          <InstallSettingsPanelContent />
        </Panel>,
        INSTALL_SETTINGS_PANEL_KEY
      )
    }, 0)
    return () => clearTimeout(timer)
  }, [panelParam, addPanel])

  return null
}
