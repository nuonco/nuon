// @ts-nocheck
'use client'

import React, { type FC, useEffect } from 'react'
import Script from 'next/script'

export const LoadEnv: FC<{ env: Record<string, string> }> = ({ env }) => {
  useEffect(() => {
    const parsedENV = JSON.parse(env)
    window.process = {
      ...window.process,
      env: {
        SEGMENT_WRITE_KEY: parsedENV.SEGMENT_WRITE_KEY,
      },
    }
  }, [])

  return <Script id="load-env">{console.log('loaded env')}</Script>
}
