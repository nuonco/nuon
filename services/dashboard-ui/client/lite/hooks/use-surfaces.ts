import { useMemo, useContext, type ReactElement } from 'react'
import { useLocation } from 'react-router'
import {
  SurfaceScopeContext,
  useSurfaceCoordinator,
} from '../providers/surfaces-provider'
import { appendSurfaceValue, type TSurfaceParam } from '../lib/surface-url'

export const useSurfaces = () => {
  const coordinator = useSurfaceCoordinator()
  const hostId = useContext(SurfaceScopeContext)

  return {
    panels: coordinator.descriptors.filter(
      (descriptor) => descriptor.type === 'panel'
    ),
    modals: coordinator.descriptors.filter(
      (descriptor) => descriptor.type === 'modal'
    ),
    openPanel: (content: ReactElement) =>
      coordinator.open('panel', hostId, content),
    openModal: (content: ReactElement) =>
      coordinator.open('modal', hostId, content),
    openPanelKey: (value: string) => coordinator.openKey('panel', value),
    openModalKey: (value: string) => coordinator.openKey('modal', value),
    closeSurface: coordinator.close,
    closeTopSurface: coordinator.closeTop,
    clearPanels: coordinator.clearPanels,
    replaceSurface: coordinator.replace,
  }
}

const useSurfaceHref = (param: TSurfaceParam, value: string) => {
  const location = useLocation()

  return useMemo(() => {
    const params = appendSurfaceValue(location.search, param, value)
    return `${location.pathname}?${params.toString()}${location.hash}`
  }, [location.hash, location.pathname, location.search, param, value])
}

export const usePanelHref = (value: string) => useSurfaceHref('panel', value)

export const useModalHref = (value: string) => useSurfaceHref('modal', value)
