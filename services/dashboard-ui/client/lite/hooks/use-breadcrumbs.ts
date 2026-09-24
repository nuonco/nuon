import { useContext, useEffect, useRef } from 'react'
import {
  BreadcrumbContext,
  type IBreadcrumbItem,
} from '../providers/breadcrumb-provider'

const useBreadcrumbContext = () => {
  const context = useContext(BreadcrumbContext)
  if (!context) {
    throw new Error('Breadcrumb hooks must be used within BreadcrumbProvider')
  }
  return context
}

export const useBreadcrumbs = (items: IBreadcrumbItem[]) => {
  const { register, unregister } = useBreadcrumbContext()
  const owner = useRef(Symbol('breadcrumb-owner'))
  const currentItems = useRef(items)
  const signature = JSON.stringify(items)
  currentItems.current = items

  useEffect(() => {
    register(owner.current, currentItems.current)
  }, [register, signature])

  useEffect(
    () => () => {
      unregister(owner.current)
    },
    [unregister]
  )
}

export const useBreadcrumbItems = () => useBreadcrumbContext().items
