'use client'

import { useEffect, useRef } from 'react'
import { useAuth } from '@/hooks/use-auth'

export const InitPylonChat = ({ PYLON_APP_ID }: { PYLON_APP_ID: string }) => {
  const { user, isLoading } = useAuth()
  const scriptLoadedRef = useRef(false)

  useEffect(() => {
    if (!PYLON_APP_ID || isLoading || !user?.email) return

    // Set identity before the widget script initializes
    window.pylon = {
      chat_settings: {
        app_id: PYLON_APP_ID,
        email: user.email,
        name: user.name || user.email,
        avatar_url: user.picture,
      },
    }

    // Only load the widget script once
    if (scriptLoadedRef.current) return
    scriptLoadedRef.current = true

    // Pylon command queue (from official docs)
    ;(function () {
      const e = window
      const t = document
      const n = function (...args: unknown[]) {
        n.e(args)
      } as Window['Pylon']
      n.q = []
      n.e = function (e: unknown[]) {
        n.q.push(e)
      }
      e.Pylon = n
      const r = function () {
        const e = t.createElement('script')
        e.setAttribute('type', 'text/javascript')
        e.setAttribute('async', 'true')
        e.setAttribute(
          'src',
          `https://widget.usepylon.com/widget/${PYLON_APP_ID}`
        )
        const n = t.getElementsByTagName('script')[0]
        if (n?.parentNode) {
          n.parentNode.insertBefore(e, n)
        } else {
          t.head.appendChild(e)
        }
      }
      if (t.readyState === 'complete') {
        r()
      } else {
        e.addEventListener('load', r, false)
      }
    })()
  }, [user, isLoading])

  return null
}
