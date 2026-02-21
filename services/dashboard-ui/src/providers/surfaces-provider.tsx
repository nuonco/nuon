'use client'

import { useRouter, usePathname } from 'next/navigation'
import React, {
  useState,
  useCallback,
  useEffect,
  type ReactNode,
} from 'react'
import { createPortal } from 'react-dom'
import { v4 as uuid } from 'uuid'
export { SurfacesContext } from '@/contexts/surfaces-context'
import { SurfacesContext, type TPanelEl, type TModalEl, type TPanels, type TModals } from '@/contexts/surfaces-context'

export function SurfacesProvider({ children }: { children: ReactNode }) {
  // Panels
  const [panels, setPanels] = useState<TPanels>([])
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    setPanels([])
  }, [pathname])

  const addPanel = useCallback(
    (content: TPanelEl, panelKey?: string, panelId?: string): string => {
      const id = panelId || uuid()
      setPanels((ps) => [
        ...ps,
        { id, key: panelKey, content, isVisible: true },
      ])
      if (panelKey) {
        const params = new URLSearchParams(window.location.search)
        params.set('panel', panelKey)
        router.replace(`?${params.toString()}`, { scroll: false })
      }
      return id
    },
    [router]
  )

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
          router.replace(`?${params.toString()}`, { scroll: false })
        }
      }, 160)
    },
    [router]
  )

  const clearPanels = useCallback(() => {
    setPanels((ps) => ps.map((p) => ({ ...p, isVisible: false })))
    setTimeout(() => {
      setPanels([])
      const params = new URLSearchParams(window.location.search)
      params.delete('panel')
      router.replace(`?${params.toString()}`, { scroll: false })
    }, 160)
  }, [router])

  // Modals
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
        router.replace(`?${params.toString()}`, { scroll: false })
      }
      return id
    },
    [router]
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
          router.replace(`?${params.toString()}`, { scroll: false })
        }
      }, 160)
    },
    [router]
  )

  return (
    <SurfacesContext.Provider
      value={{
        panels,
        modals,
        addPanel,
        clearPanels,
        removePanel,
        addModal,
        removeModal,
      }}
    >
      {children}
      {/* Panels */}
      {panels.map((p) => (
        <React.Fragment key={p.id}>
          {React.isValidElement(p.content)
            ? createPortal(
                React.cloneElement(p.content, {
                  panelId: p.id,
                  panelKey: p?.key,
                  isVisible: p.isVisible,
                }),
                document.getElementById('panel-root')
              )
            : null}
        </React.Fragment>
      ))}
      <div id="panel-root" />
      {/* Modals */}
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
