import {
  createContext,
  useCallback,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'

interface IStatusBarContext {
  content: ReactNode
  register: (owner: symbol, content: ReactNode) => void
  unregister: (owner: symbol) => void
}

export const StatusBarContext = createContext<IStatusBarContext | null>(null)

export const StatusBarProvider = ({ children }: { children: ReactNode }) => {
  const activeOwner = useRef<symbol | undefined>(undefined)
  const [content, setContent] = useState<ReactNode>(null)

  const register = useCallback((owner: symbol, next: ReactNode) => {
    activeOwner.current = owner
    setContent(next)
  }, [])

  const unregister = useCallback((owner: symbol) => {
    if (activeOwner.current !== owner) return
    activeOwner.current = undefined
    setContent(null)
  }, [])

  const value = useMemo(
    () => ({ content, register, unregister }),
    [content, register, unregister]
  )

  return (
    <StatusBarContext.Provider value={value}>
      {children}
    </StatusBarContext.Provider>
  )
}
