import { useEffect, useRef } from 'react'
import { useNavigate, useParams } from 'react-router'

const CHORD_TIMEOUT_MS = 1500

const ORG_DESTINATIONS: Record<string, string> = {
  d: '',
  a: '/apps',
  i: '/installs',
  t: '/team',
  r: '/runner',
  w: '/webhooks',
}

export function useNavShortcuts() {
  const navigate = useNavigate()
  const { orgId } = useParams<{ orgId: string }>()
  const orgIdRef = useRef(orgId)

  useEffect(() => {
    orgIdRef.current = orgId
  }, [orgId])

  useEffect(() => {
    let chordPrimed = false
    let timer: ReturnType<typeof setTimeout> | null = null

    const reset = () => {
      chordPrimed = false
      if (timer) {
        clearTimeout(timer)
        timer = null
      }
    }

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey) return

      const target = e.target as HTMLElement | null
      const tag = target?.tagName
      if (
        tag === 'INPUT' ||
        tag === 'TEXTAREA' ||
        tag === 'SELECT' ||
        target?.isContentEditable
      ) {
        return
      }

      const key = e.key.toLowerCase()

      if (chordPrimed) {
        const path = ORG_DESTINATIONS[key]
        if (path !== undefined) {
          e.preventDefault()
          const id = orgIdRef.current
          navigate(id ? `/${id}${path}` : '/')
        }
        reset()
        return
      }

      if (key === 'g' && !e.shiftKey) {
        chordPrimed = true
        timer = setTimeout(reset, CHORD_TIMEOUT_MS)
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      reset()
    }
  }, [navigate])
}
