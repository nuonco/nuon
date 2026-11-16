import { createContext, useContext } from 'react'

export const PanelVisibilityContext = createContext<boolean>(true)

export const usePanelVisibility = () => useContext(PanelVisibilityContext)
