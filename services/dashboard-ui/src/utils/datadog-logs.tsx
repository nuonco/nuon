'use client'

import { datadogLogs } from '@datadog/browser-logs'
import React, { type FC, useEffect } from 'react'

export const InitDatadogLogs: FC = () => {
  useEffect(() => {
    const initDDLogs = () => {
      datadogLogs.init({
        clientToken: process?.env?.NEXT_PUBLIC_DATADOG_CLIENT_TOKEN,
        site: process?.env?.NEXT_PUBLIC_DATADOG_SITE,
        forwardConsoleLogs: ['error', 'info'],
        forwardErrorsToLogs: true,
        sessionSampleRate: 100,
        env: process?.env?.NEXT_PUBLIC_DATADOG_ENV,
        service: 'dashboard',
      })
    }

    initDDLogs()
  }, [])

  return <></>
}
