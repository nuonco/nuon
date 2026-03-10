import { createContext, useState, useEffect, type ReactNode } from 'react'

type PageTitleContextValue = {
  title: string
  updateTitle: (title: string) => void
}

export const PageTitleContext = createContext<PageTitleContextValue | undefined>(undefined)

export function PageTitleProvider({ children }: { children: ReactNode }) {
  const [title, setTitle] = useState('')

  useEffect(() => {
    document.title = title ? `${title} | Nuon` : 'Nuon'
  }, [title])

  return (
    <PageTitleContext.Provider value={{ title, updateTitle: setTitle }}>
      {children}
    </PageTitleContext.Provider>
  )
}
