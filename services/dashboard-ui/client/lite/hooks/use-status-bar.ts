import { useContext, useEffect, useRef, type ReactNode } from 'react'
import { StatusBarContext } from '../providers/status-bar-provider'

const useStatusBarContext = () => {
  const context = useContext(StatusBarContext)
  if (!context) {
    throw new Error('Status bar hooks must be used within StatusBarProvider')
  }
  return context
}

export const useStatusBar = (content: ReactNode, deps: unknown[]) => {
  const { register, unregister } = useStatusBarContext()
  const owner = useRef(Symbol('status-bar-owner'))
  const currentContent = useRef(content)
  const signature = JSON.stringify(deps)
  currentContent.current = content

  useEffect(() => {
    register(owner.current, currentContent.current)
  }, [register, signature])

  useEffect(
    () => () => {
      unregister(owner.current)
    },
    [unregister]
  )
}

export const useStatusBarContent = () => useStatusBarContext().content
