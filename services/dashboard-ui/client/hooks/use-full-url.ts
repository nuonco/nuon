'use client'

import { usePathname, useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'

export function useFullUrl() {
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const [fullUrl, setFullUrl] = useState<string>('')

  useEffect(() => {
    const url = new URL(pathname, window.location.origin)

    if (searchParams) {
      searchParams.forEach((value, key) => {
        url.searchParams.set(key, value)
      })
    }

    setFullUrl(url.toString())
  }, [pathname, searchParams])

  return fullUrl
}
