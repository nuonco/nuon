import { createContext, useContext } from 'react'

export type TConfigDiffFocus = {
  sectionKey: string
  entityName?: string
  nonce: number
}

type TConfigDiffFocusValue = {
  requestFocus: (sectionKey: string, entityName?: string) => void
}

export const ConfigDiffFocusContext = createContext<
  TConfigDiffFocusValue | undefined
>(undefined)

export function useConfigDiffFocus() {
  return useContext(ConfigDiffFocusContext)
}
