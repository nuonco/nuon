import { useLocation } from 'react-router'

export function useFullUrl() {
  const { pathname, search } = useLocation()
  return `${window.location.origin}${pathname}${search}`
}
