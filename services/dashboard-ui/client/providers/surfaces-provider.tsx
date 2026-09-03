import { useNavigate, useLocation } from 'react-router'
import React, {
  createContext,
  useState,
  useCallback,
  useEffect,
  type ReactElement,
  type ReactNode,
} from 'react'
import { createPortal } from 'react-dom'
import { type IPanel } from '@/components/surfaces/Panel'
import { type IModal } from '@/components/surfaces/Modal'

const uuid = () => crypto.randomUUID()

type TPanelEl = ReactElement<IPanel & { ref?: React.Ref<HTMLDivElement> }>
type TPanels = {
  id: string
  key?: string
  content: TPanelEl
  isVisible: boolean
  pathname: string
}[]

type TModalEl = ReactElement<IModal & { ref?: React.Ref<HTMLDivElement> }>
type TModals = {
  id: string
  key?: string
  content: TModalEl
  isVisible: boolean
}[]

type TSurfacesContext = {
  panels: TPanels
  modals: TModals
  addPanel: (content: TPanelEl, panelKey?: string, panelId?: string) => string
  updatePanel: (id: string, content: TPanelEl) => void
  clearPanels: () => void
  removePanel: (id: string, panelKey?: string) => void
  addModal: (content: TModalEl, modalKey?: string) => string
  removeModal: (id: string, modalKey?: string) => void
}

export const SurfacesContext = createContext<TSurfacesContext | undefined>(
  undefined
)

export function SurfacesProvider({ children }: { children: ReactNode }) {
  const [panels, setPanels] = useState<TPanels>([])
  const navigate = useNavigate()
  const { pathname } = useLocation()

  // Panels are keyed to the route that opened them. A route change drops the
  // previous route's panels, but not ones a newly mounted child opened for the
  // route being navigated to — child effects run before this one.
  useEffect(() => {
    setPanels((ps) => ps.filter((p) => p.pathname === pathname))
  }, [pathname])

  const addPanel = useCallback(
    (content: TPanelEl, panelKey?: string, panelId?: string): string => {
      const id = panelId || uuid()
      setPanels((ps) => [
        ...ps,
        {
          id,
          key: panelKey,
          content,
          isVisible: true,
          pathname: window.location.pathname,
        },
      ])
      if (panelKey) {
        const params = new URLSearchParams(window.location.search)
        params.set('panel', panelKey)
        navigate(`?${params.toString()}`, { replace: true })
      }
      return id
    },
    [navigate]
  )

  const updatePanel = useCallback((id: string, content: TPanelEl) => {
    setPanels((ps) => ps.map((p) => (p.id === id ? { ...p, content } : p)))
  }, [])

  const removePanel = useCallback(
    (id: string, panelKey?: string) => {
      setPanels((ps) =>
        ps.map((p) => (p?.id === id ? { ...p, isVisible: false } : p))
      )
      setTimeout(() => {
        setPanels((ps) => ps.filter((p) => p?.id !== id))
        if (panelKey) {
          const params = new URLSearchParams(window.location.search)
          params.delete('panel')
          navigate(`?${params.toString()}`, { replace: true })
        }
      }, 160)
    },
    [navigate]
  )

  const clearPanels = useCallback(() => {
    setPanels((ps) => ps.map((p) => ({ ...p, isVisible: false })))
    setTimeout(() => {
      setPanels([])
      const params = new URLSearchParams(window.location.search)
      params.delete('panel')
      navigate(`?${params.toString()}`, { replace: true })
    }, 160)
  }, [navigate])

  const [modals, setModals] = useState<TModals>([])

  const addModal = useCallback(
    (content: TModalEl, modalKey?: string): string => {
      const id = uuid()
      setModals((ms) => [
        ...ms,
        { id, key: modalKey, content, isVisible: true },
      ])
      if (modalKey) {
        const params = new URLSearchParams(window.location.search)
        params.set('modal', modalKey)
        navigate(`?${params.toString()}`, { replace: true })
      }
      return id
    },
    [navigate]
  )

  const removeModal = useCallback(
    (id: string, modalKey?: string) => {
      setModals((ms) =>
        ms.map((m) => (m?.id === id ? { ...m, isVisible: false } : m))
      )
      setTimeout(() => {
        setModals((ms) => ms.filter((m) => m?.id !== id))
        if (modalKey) {
          const params = new URLSearchParams(window.location.search)
          params.delete('modal')
          navigate(`?${params.toString()}`, { replace: true })
        }
      }, 160)
    },
    [navigate]
  )

  return (
    <SurfacesContext.Provider
      value={{
        panels,
        modals,
        addPanel,
        updatePanel,
        clearPanels,
        removePanel,
        addModal,
        removeModal,
      }}
    >
      {children}
      {panels.map((p) => (
        <React.Fragment key={p.id}>
          {React.isValidElement(p.content)
            ? createPortal(
                React.cloneElement(p.content, {
                  panelId: p.id,
                  panelKey: p?.key,
                  isVisible: p.isVisible,
                }),
                document.getElementById('panel-root')!
              )
            : null}
        </React.Fragment>
      ))}
      <div id="panel-root" />
      {modals.map((m) => (
        <React.Fragment key={m.id}>
          {React.isValidElement(m.content)
            ? createPortal(
                React.cloneElement(m.content, {
                  modalId: m.id,
                  modalKey: m?.key,
                  isVisible: m.isVisible,
                }),
                document.getElementById('modal-root')
              )
            : null}
        </React.Fragment>
      ))}
      <div id="modal-root" />
    </SurfacesContext.Provider>
  )
}
