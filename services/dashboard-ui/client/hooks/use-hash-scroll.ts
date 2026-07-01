import { useEffect } from 'react'
import { useLocation } from 'react-router'
import { scrollElementIntoView } from '@/utils/scroll'

export function useHashScroll() {
  const location = useLocation()

  useEffect(() => {
    if (location.hash) {
      setTimeout(() => {
        const id = location.hash.replace('#', '')
        scrollElementIntoView(document.getElementById(id), { block: 'start' })
      }, 0)
    }
  }, [location])
}
