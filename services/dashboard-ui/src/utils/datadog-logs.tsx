'use client'

import { datadogLogs } from '@datadog/browser-logs'
import React, { type FC, useEffect } from 'react'

export const InitDatadogLogs: FC = () => {
   const ddEnv =
        process?.env?.NEXT_PUBLIC_DATADOG_ENV ||
        window.location?.host?.split('.')[1]
  
  useEffect(() => {
    const initDDLogs = () => {
      datadogLogs.init({
        clientToken: process?.env?.NEXT_PUBLIC_DATADOG_CLIENT_TOKEN ||
          'pub6fb6cfe0d2ec271a2456660e54ba5e08',
        site: process?.env?.NEXT_PUBLIC_DATADOG_SITE || 'us5.datadoghq.com',
        forwardConsoleLogs: ['error', 'info'],
        forwardErrorsToLogs: true,
        sessionSampleRate: 100,
        env: ddEnv || 'local',
        service: 'dashboard',
      })
    }

    initDDLogs()
  }, [])

  return <></>
}
