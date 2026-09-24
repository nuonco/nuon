import { useContext, useEffect, useRef } from 'react'
import {
  composePageTitle,
  PageTitleContext,
  type TPageTitleSegment,
} from '../providers/page-title-provider'

const usePageTitleContext = () => {
  const context = useContext(PageTitleContext)
  if (!context) {
    throw new Error('Page title hooks must be used within PageTitleProvider')
  }
  return context
}

export const usePageTitle = (...segments: TPageTitleSegment[]) => {
  const { register, unregister } = usePageTitleContext()
  const owner = useRef(Symbol('page-title-owner'))
  const title = composePageTitle(segments)

  useEffect(() => {
    register(owner.current, title)
  }, [register, title])

  useEffect(
    () => () => {
      unregister(owner.current)
    },
    [unregister]
  )
}

export const usePageTitleValue = () => usePageTitleContext().title
