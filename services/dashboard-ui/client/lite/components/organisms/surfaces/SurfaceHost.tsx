import {
  createContext,
  useContext,
  useId,
  useLayoutEffect,
  type ReactNode,
} from 'react'
import { createPortal } from 'react-dom'
import {
  SurfaceScopeContext,
  useSurfaceCoordinator,
  type ISurfaceRegistration,
  type TSurfaceType,
} from '../../../providers/surfaces-provider'

interface ISurfaceInstanceContext {
  id: string
  type: TSurfaceType
  visible: boolean
  topmost: boolean
  coveredBy: number
  zIndex: number
  close: () => void
  replace: (content: React.ReactElement) => void
}

const SurfaceInstanceContext = createContext<ISurfaceInstanceContext | null>(
  null
)

const EMPTY_REGISTRATIONS: ISurfaceRegistration[] = []

export interface ISurfaceHost {
  scope: string
  registrations?: ISurfaceRegistration[]
  children: ReactNode
}

export const SurfaceHost = ({
  scope,
  registrations = EMPTY_REGISTRATIONS,
  children,
}: ISurfaceHost) => {
  const reactId = useId()
  const hostId = `${scope}:${reactId}`
  const coordinator = useSurfaceCoordinator()

  useLayoutEffect(() => {
    coordinator.register(hostId, registrations)
  }, [coordinator.register, hostId, registrations])

  useLayoutEffect(
    () => () => coordinator.unregister(hostId),
    [coordinator.unregister, hostId]
  )

  const topmostId = coordinator.descriptors.at(-1)?.id
  const owned = coordinator.descriptors.filter(
    (descriptor) => descriptor.hostId === hostId
  )

  return (
    <SurfaceScopeContext.Provider value={hostId}>
      {children}
      {coordinator.portalRoot
        ? owned.map((descriptor) => {
            const index = coordinator.descriptors.findIndex(
              (item) => item.id === descriptor.id
            )
            const coveredBy = coordinator.descriptors
              .slice(index + 1)
              .filter((item) => item.type === descriptor.type).length
            const instance = {
              id: descriptor.id,
              type: descriptor.type,
              visible: descriptor.visible,
              topmost: descriptor.id === topmostId,
              coveredBy,
              zIndex: 100 + index * 10,
              close: () => coordinator.close(descriptor.id),
              replace: (content: React.ReactElement) =>
                coordinator.replace(descriptor.id, content),
            }

            return createPortal(
              <SurfaceInstanceContext.Provider
                key={descriptor.id}
                value={instance}
              >
                {descriptor.content}
              </SurfaceInstanceContext.Provider>,
              coordinator.portalRoot
            )
          })
        : null}
    </SurfaceScopeContext.Provider>
  )
}

export const useSurfaceInstance = () => {
  const context = useContext(SurfaceInstanceContext)
  if (!context) {
    throw new Error('Modal and Panel must be opened through a SurfaceHost')
  }
  return context
}

export const useCurrentSurface = () => {
  const { close, id, replace, type } = useSurfaceInstance()
  return { close, id, replace, type }
}
