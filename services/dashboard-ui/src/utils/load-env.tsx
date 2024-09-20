// @ts-nocheck
'use client'

import React, { type FC, useEffect } from 'react'
import Script from 'next/script'
import { AnalyticsBrowser } from '@segment/analytics-next'

export const LoadEnv: FC<{ env: Record<string, string> }> = ({ env }) => {
  useEffect(() => {
    const parsedENV = JSON.parse(env)
    
    window.analytics = AnalyticsBrowser.load({
      writeKey: parsedENV.SEGMENT_WRITE_KEY!,
    })
  }, [])

  return <Script id="load-env">{console.log('analytics initialized')}</Script>
}
