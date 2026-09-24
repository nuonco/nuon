import {
  createContext,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'

export const APP_TITLE = 'Nuon'

export type TPageTitleSegment = string | undefined | null | false

export const composePageTitle = (segments: TPageTitleSegment[]) =>
  segments
    .filter(
      (segment): segment is string =>
        typeof segment === 'string' && segment.length > 0
    )
    .join(' | ')

export const documentTitleFor = (title: string) =>
  title ? `${title} | ${APP_TITLE}` : APP_TITLE

interface IPageTitleContext {
  title: string
  register: (owner: symbol, title: string) => void
  unregister: (owner: symbol) => void
}

export const PageTitleContext = createContext<IPageTitleContext | null>(null)

export const PageTitleProvider = ({ children }: { children: ReactNode }) => {
  const activeOwner = useRef<symbol | undefined>(undefined)
  const [title, setTitle] = useState('')

  const register = useCallback((owner: symbol, nextTitle: string) => {
    activeOwner.current = owner
    setTitle(nextTitle)
  }, [])

  const unregister = useCallback((owner: symbol) => {
    if (activeOwner.current !== owner) return
    activeOwner.current = undefined
    setTitle('')
  }, [])

  useEffect(() => {
    document.title = documentTitleFor(title)
  }, [title])

  const value = useMemo(
    () => ({ title, register, unregister }),
    [register, title, unregister]
  )

  return (
    <PageTitleContext.Provider value={value}>
      {children}
    </PageTitleContext.Provider>
  )
}
