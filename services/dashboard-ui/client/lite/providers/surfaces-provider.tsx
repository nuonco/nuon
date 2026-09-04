import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type ReactElement,
  type ReactNode,
} from 'react'
import { useLocation, useNavigate, useSearchParams } from 'react-router'
import {
  appendSurfaceValue,
  parseSurfaceValue,
  truncateSurfaceValues,
  type ISurfaceValue,
  type TSurfaceParam,
} from '../lib/surface-url'
import { SURFACE_TRANSITION_MS } from '../lib/surface-motion'

export type TSurfaceType = 'panel' | 'modal'

export interface ISurfaceRegistration {
  key: string
  type: TSurfaceType
  render: (surface: ISurfaceValue) => ReactElement
}

export interface ISurfaceDescriptor {
  id: string
  type: TSurfaceType
  hostId: string
  content: ReactElement
  visible: boolean
  order: number
  source: 'imperative' | 'url'
  opener?: HTMLElement
  urlIndex?: number
  value?: string
}

interface IRegisteredSurface extends ISurfaceRegistration {
  hostId: string
}

interface ISurfacesContext {
  descriptors: ISurfaceDescriptor[]
  portalRoot: HTMLElement | null
  register: (hostId: string, registrations: ISurfaceRegistration[]) => void
  unregister: (hostId: string) => void
  open: (
    type: TSurfaceType,
    hostId: string | null,
    content: ReactElement
  ) => string
  openKey: (type: TSurfaceType, value: string) => void
  close: (id: string) => void
  closeTop: () => void
  clearPanels: () => void
  replace: (id: string, content: ReactElement) => void
}

export const SurfacesContext = createContext<ISurfacesContext | null>(null)
export const SurfaceScopeContext = createContext<string | null>(null)

const surfaceParam = (type: TSurfaceType): TSurfaceParam =>
  type === 'panel' ? 'panel' : 'modal'

const urlSurfaceId = (type: TSurfaceType, index: number, value: string) =>
  `url:${type}:${index}:${value}`

const activeElement = () =>
  document.activeElement instanceof HTMLElement
    ? document.activeElement
    : undefined

export const panelRegistration = (
  key: string,
  render: (surface: ISurfaceValue) => ReactElement
): ISurfaceRegistration => ({ key, type: 'panel', render })

export const modalRegistration = (
  key: string,
  render: (surface: ISurfaceValue) => ReactElement
): ISurfaceRegistration => ({ key, type: 'modal', render })

export const SurfacesProvider = ({ children }: { children: ReactNode }) => {
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams] = useSearchParams()
  const [portalRoot, setPortalRoot] = useState<HTMLElement | null>(null)
  const [registrations, setRegistrations] = useState<IRegisteredSurface[]>([])
  const [urlPanels, setUrlPanels] = useState<ISurfaceDescriptor[]>([])
  const [urlModals, setUrlModals] = useState<ISurfaceDescriptor[]>([])
  const [imperative, setImperative] = useState<ISurfaceDescriptor[]>([])
  const timers = useRef(new Map<string, ReturnType<typeof setTimeout>>())
  const ownedUrlEntries = useRef(new Set<string>())
  const urlOpeners = useRef(new Map<string, HTMLElement | undefined>())
  const urlOrders = useRef(new Map<string, number>())
  const nextOrder = useRef(0)
  const latestSearch = useRef(location.search)
  const appRootRef = useRef<HTMLDivElement>(null)
  latestSearch.current = location.search

  const restoreAndRemove = useCallback(
    (descriptor: ISurfaceDescriptor, remove: (id: string) => void) => {
      const existing = timers.current.get(descriptor.id)
      if (existing) clearTimeout(existing)
      const timer = setTimeout(() => {
        remove(descriptor.id)
        timers.current.delete(descriptor.id)
        requestAnimationFrame(() => {
          if (descriptor.opener?.isConnected) descriptor.opener.focus()
        })
      }, SURFACE_TRANSITION_MS)
      timers.current.set(descriptor.id, timer)
    },
    []
  )

  const register = useCallback(
    (hostId: string, next: ISurfaceRegistration[]) => {
      const registered = next.map((registration) => ({
        ...registration,
        hostId,
      }))
      setRegistrations((current) => [
        ...current.filter((item) => item.hostId !== hostId),
        ...registered,
      ])
    },
    []
  )

  const unregister = useCallback((hostId: string) => {
    const removeHostDescriptors = (
      current: ISurfaceDescriptor[]
    ): ISurfaceDescriptor[] => {
      current
        .filter((item) => item.hostId === hostId)
        .forEach((item) => {
          const timer = timers.current.get(item.id)
          if (timer) clearTimeout(timer)
          timers.current.delete(item.id)
        })
      return current.filter((item) => item.hostId !== hostId)
    }
    setRegistrations((current) =>
      current.filter((item) => item.hostId !== hostId)
    )
    setUrlPanels(removeHostDescriptors)
    setUrlModals(removeHostDescriptors)
    setImperative(removeHostDescriptors)
  }, [])

  const reconcileUrlStack = useCallback(
    (
      type: TSurfaceType,
      values: string[],
      current: ISurfaceDescriptor[],
      remove: (id: string) => void
    ) => {
      values.forEach((value, index) => {
        const id = urlSurfaceId(type, index, value)
        if (!urlOrders.current.has(id)) {
          nextOrder.current += 1
          urlOrders.current.set(id, nextOrder.current)
        }
      })
      const resolved = values.flatMap((value, index) => {
        const parsed = parseSurfaceValue(value)
        const matches = registrations.filter(
          (registration) =>
            registration.type === type && registration.key === parsed.key
        )
        if (matches.length !== 1) {
          if (matches.length > 1 && process.env.NODE_ENV === 'development') {
            console.warn(
              `Surface registration "${parsed.key}" is duplicated in the active route.`
            )
          }
          return []
        }

        const id = urlSurfaceId(type, index, value)
        const timer = timers.current.get(id)
        if (timer) {
          clearTimeout(timer)
          timers.current.delete(id)
        }
        const existing = current.find((descriptor) => descriptor.id === id)
        return [
          {
            ...existing,
            id,
            type,
            hostId: matches[0].hostId,
            content: matches[0].render(parsed),
            visible: true,
            order: urlOrders.current.get(id) ?? 0,
            source: 'url' as const,
            opener: existing?.opener ?? urlOpeners.current.get(id),
            urlIndex: index,
            value,
          },
        ]
      })
      const ids = new Set(resolved.map((descriptor) => descriptor.id))
      const exiting = current
        .filter((descriptor) => !ids.has(descriptor.id))
        .map((descriptor) => {
          if (descriptor.visible) restoreAndRemove(descriptor, remove)
          return { ...descriptor, visible: false }
        })

      return [...resolved, ...exiting]
    },
    [registrations, restoreAndRemove]
  )

  const panelValues = searchParams.getAll('panel')
  const modalValues = searchParams.getAll('modal')
  const panelSignature = panelValues.join('\u0000')
  const modalSignature = modalValues.join('\u0000')

  useEffect(() => {
    setUrlPanels((current) =>
      reconcileUrlStack('panel', panelValues, current, (id) =>
        setUrlPanels((items) => items.filter((item) => item.id !== id))
      )
    )
  }, [panelSignature, reconcileUrlStack])

  useEffect(() => {
    setUrlModals((current) =>
      reconcileUrlStack('modal', modalValues, current, (id) =>
        setUrlModals((items) => items.filter((item) => item.id !== id))
      )
    )
  }, [modalSignature, reconcileUrlStack])

  const open = useCallback(
    (type: TSurfaceType, hostId: string | null, content: ReactElement) => {
      if (!hostId) {
        throw new Error('A SurfaceHost is required to open a surface')
      }
      const id = crypto.randomUUID()
      nextOrder.current += 1
      const order = nextOrder.current
      setImperative((current) => [
        ...current,
        {
          id,
          type,
          hostId,
          content,
          visible: true,
          order,
          source: 'imperative',
          opener: activeElement(),
        },
      ])
      return id
    },
    []
  )

  const openKey = useCallback(
    (type: TSurfaceType, value: string) => {
      const param = surfaceParam(type)
      const params = appendSurfaceValue(latestSearch.current, param, value)
      const index = params.getAll(param).length - 1
      const id = urlSurfaceId(type, index, value)
      nextOrder.current += 1
      ownedUrlEntries.current.add(id)
      urlOpeners.current.set(id, activeElement())
      urlOrders.current.set(id, nextOrder.current)
      latestSearch.current = `?${params.toString()}`
      navigate(
        {
          pathname: location.pathname,
          search: `?${params.toString()}`,
          hash: location.hash,
        },
        { replace: false }
      )
    },
    [location.hash, location.pathname, navigate]
  )

  const close = useCallback(
    (id: string) => {
      const descriptor = [...urlPanels, ...urlModals, ...imperative].find(
        (item) => item.id === id
      )
      if (!descriptor) return

      if (descriptor.source === 'imperative') {
        setImperative((current) =>
          current.map((item) =>
            item.id === id ? { ...item, visible: false } : item
          )
        )
        restoreAndRemove(descriptor, (surfaceId) =>
          setImperative((current) =>
            current.filter((item) => item.id !== surfaceId)
          )
        )
        return
      }

      const param = surfaceParam(descriptor.type)
      const values = new URLSearchParams(location.search).getAll(param)
      const isTop = descriptor.urlIndex === values.length - 1
      if (isTop && ownedUrlEntries.current.has(id)) {
        ownedUrlEntries.current.delete(id)
        navigate(-1)
        return
      }

      const params = truncateSurfaceValues(
        location.search,
        param,
        descriptor.urlIndex ?? values.length - 1
      )
      navigate(
        {
          pathname: location.pathname,
          search: params.size ? `?${params.toString()}` : '',
          hash: location.hash,
        },
        { replace: true }
      )
    },
    [
      imperative,
      location.hash,
      location.pathname,
      location.search,
      navigate,
      restoreAndRemove,
      urlModals,
      urlPanels,
    ]
  )

  const panels = [
    ...urlPanels,
    ...imperative.filter((item) => item.type === 'panel'),
  ].sort((a, b) => a.order - b.order)
  const modals = [
    ...urlModals,
    ...imperative.filter((item) => item.type === 'modal'),
  ].sort((a, b) => a.order - b.order)
  const descriptors = [...panels, ...modals]
  const top = modals.at(-1) ?? panels.at(-1)

  const closeTop = useCallback(() => {
    if (top) close(top.id)
  }, [close, top])

  const clearPanels = useCallback(() => {
    imperative
      .filter((item) => item.type === 'panel' && item.visible)
      .forEach((item) => close(item.id))
    const params = new URLSearchParams(location.search)
    params.delete('panel')
    navigate(
      {
        pathname: location.pathname,
        search: params.size ? `?${params.toString()}` : '',
        hash: location.hash,
      },
      { replace: true }
    )
  }, [
    close,
    imperative,
    location.hash,
    location.pathname,
    location.search,
    navigate,
  ])

  const replace = useCallback((id: string, content: ReactElement) => {
    setImperative((current) =>
      current.map((item) => (item.id === id ? { ...item, content } : item))
    )
  }, [])

  const hasSurfaces = descriptors.length > 0
  useLayoutEffect(() => {
    const appRoot = appRootRef.current
    if (appRoot) appRoot.inert = hasSurfaces
    if (!hasSurfaces) return
    const overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.body.style.overflow = overflow
      if (appRoot) appRoot.inert = false
    }
  }, [hasSurfaces])

  useEffect(
    () => () => {
      timers.current.forEach(clearTimeout)
    },
    []
  )

  const value = useMemo(
    () => ({
      descriptors,
      portalRoot,
      register,
      unregister,
      open,
      openKey,
      close,
      closeTop,
      clearPanels,
      replace,
    }),
    [
      clearPanels,
      close,
      closeTop,
      descriptors,
      open,
      openKey,
      portalRoot,
      register,
      replace,
      unregister,
    ]
  )

  return (
    <SurfacesContext.Provider value={value}>
      <div ref={appRootRef} className="contents">
        {children}
      </div>
      <div ref={setPortalRoot} id="lite-surface-root" />
    </SurfacesContext.Provider>
  )
}

export const useSurfaceCoordinator = () => {
  const context = useContext(SurfacesContext)
  if (!context) {
    throw new Error('Surface components must be inside SurfacesProvider')
  }
  return context
}
