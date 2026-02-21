import { createContext, type ReactElement } from 'react'
import { type IPanel } from '@/components/surfaces/Panel'
import { type IModal } from '@/components/surfaces/Modal'

export type TPanelEl = ReactElement<IPanel & { ref?: React.Ref<HTMLDivElement> }>
export type TModalEl = ReactElement<IModal & { ref?: React.Ref<HTMLDivElement> }>

export type TPanels = {
  id: string
  key?: string
  content: TPanelEl
  isVisible: boolean
}[]

export type TModals = {
  id: string
  key?: string
  content: TModalEl
  isVisible: boolean
}[]

type TSurfacesContext = {
  panels: TPanels
  modals: TModals
  addPanel: (content: TPanelEl, panelKey?: string, panelId?: string) => string
  clearPanels: () => void
  removePanel: (id: string, panelKey?: string) => void
  addModal: (content: TModalEl, modalKey?: string) => string
  removeModal: (id: string, modalKey?: string) => void
}

export const SurfacesContext = createContext<TSurfacesContext | undefined>(
  undefined
)
