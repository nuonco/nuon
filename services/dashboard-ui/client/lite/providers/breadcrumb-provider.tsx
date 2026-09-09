import {
  createContext,
  useCallback,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'

export interface IBreadcrumbItem {
  label?: string
  href?: string
  loadingWidth?: number
}

interface IBreadcrumbContext {
  items: IBreadcrumbItem[]
  register: (owner: symbol, items: IBreadcrumbItem[]) => void
  unregister: (owner: symbol) => void
}

export const BreadcrumbContext = createContext<IBreadcrumbContext | null>(null)

export const BreadcrumbProvider = ({ children }: { children: ReactNode }) => {
  const activeOwner = useRef<symbol | undefined>(undefined)
  const [items, setItems] = useState<IBreadcrumbItem[]>([])

  const register = useCallback((owner: symbol, nextItems: IBreadcrumbItem[]) => {
    activeOwner.current = owner
    setItems(nextItems)
  }, [])

  const unregister = useCallback((owner: symbol) => {
    if (activeOwner.current !== owner) return
    activeOwner.current = undefined
    setItems([])
  }, [])

  const value = useMemo(
    () => ({ items, register, unregister }),
    [items, register, unregister]
  )

  return (
    <BreadcrumbContext.Provider value={value}>
      {children}
    </BreadcrumbContext.Provider>
  )
}
