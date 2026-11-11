import { useLocation } from 'react-router'
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
import { useReplaceSearchParams } from '@/hooks/use-replace-search-params'

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

const PANEL_PARAM = 'panel'

// Repeated params rather than one, since panels stack and a deep link needs
// to reconstruct the whole chain, not just the topmost panel.
export const panelKeyStack = (search: URLSearchParams | string) =>
  (typeof search === 'string'
    ? new URLSearchParams(search)
    : search
  ).getAll(PANEL_PARAM)

export const hasPanelKey = (search: URLSearchParams | string, key: string) =>
  panelKeyStack(search).includes(key)

export const withPanelKey = (params: URLSearchParams, key: string) => {
  if (!params.getAll(PANEL_PARAM).includes(key)) params.append(PANEL_PARAM, key)
  return params
}

export const withoutPanelKey = (params: URLSearchParams, key: string) => {
  const remaining = params.getAll(PANEL_PARAM).filter((k) => k !== key)
  params.delete(PANEL_PARAM)
  remaining.forEach((k) => params.append(PANEL_PARAM, k))
  return params
}

export function SurfacesProvider({ children }: { children: ReactNode }) {
  const [panels, setPanels] = useState<TPanels>([])
  const { pathname } = useLocation()
  const replaceSearch = useReplaceSearchParams()
  const updateSearch = useCallback(
    (update: (params: URLSearchParams) => URLSearchParams) => {
      replaceSearch((params) => {
        update(params)
      })
    },
    [replaceSearch]
  )

  useEffect(() => {
    setPanels((ps) => ps.filter((p) => p.pathname === pathname))
  }, [pathname])

  const addPanel = useCallback(
    (content: TPanelEl, panelKey?: string, panelId?: string): string => {
      const id = panelId || uuid()
      const order = panelKeyStack(window.location.search)
      setPanels((ps) => {
        if (panelKey && ps.some((p) => p.key === panelKey && p.isVisible)) {
          return ps
        }
        const withNew = [
          ...ps,
          {
            id,
            key: panelKey,
            content,
            isVisible: true,
            pathname,
          },
        ]
        return withNew
          .map((panel, index) => ({ panel, index }))
          .sort((a, b) => {
            const rank = (panel: (typeof withNew)[number], index: number) => {
              const stackIndex = panel.key ? order.indexOf(panel.key) : -1
              return stackIndex === -1 ? order.length + index : stackIndex
            }
            return rank(a.panel, a.index) - rank(b.panel, b.index)
          })
          .map(({ panel }) => panel)
      })
      if (panelKey) {
        updateSearch((params) => withPanelKey(params, panelKey))
      }
      return id
    },
    [pathname, updateSearch]
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
          updateSearch((params) => withoutPanelKey(params, panelKey))
        }
      }, 160)
    },
    [updateSearch]
  )

  const clearPanels = useCallback(() => {
    setPanels((ps) => ps.map((p) => ({ ...p, isVisible: false })))
    setTimeout(() => {
      setPanels([])
      updateSearch((params) => {
        params.delete(PANEL_PARAM)
        return params
      })
    }, 160)
  }, [updateSearch])

  const [modals, setModals] = useState<TModals>([])

  const addModal = useCallback(
    (content: TModalEl, modalKey?: string): string => {
      const id = uuid()
      setModals((ms) => [
        ...ms,
        { id, key: modalKey, content, isVisible: true },
      ])
      if (modalKey) {
        updateSearch((params) => {
          params.set('modal', modalKey)
          return params
        })
      }
      return id
    },
    [updateSearch]
  )

  const removeModal = useCallback(
    (id: string, modalKey?: string) => {
      setModals((ms) =>
        ms.map((m) => (m?.id === id ? { ...m, isVisible: false } : m))
      )
      setTimeout(() => {
        setModals((ms) => ms.filter((m) => m?.id !== id))
        if (modalKey) {
          updateSearch((params) => {
            params.delete('modal')
            return params
          })
        }
      }, 160)
    },
    [updateSearch]
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
      {panels.length > 0
        ? createPortal(
            panels.map((p) =>
              React.isValidElement(p.content)
                ? React.cloneElement(p.content, {
                    key: p.id,
                    panelId: p.id,
                    panelKey: p?.key,
                    isVisible: p.isVisible,
                  })
                : null
            ),
            document.getElementById('panel-root')!
          )
        : null}
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
