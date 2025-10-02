'use client'

import { datadogLogs } from '@datadog/browser-logs'
import { type FC, useEffect } from 'react'

let isDatadogInitialized = false

const initDatadogLogs = (env: 'local' | 'stage' | 'prod') => {
  if (isDatadogInitialized) return

  datadogLogs.init({
    clientToken:
      process?.env?.NEXT_PUBLIC_DATADOG_CLIENT_TOKEN ||
      'pub6fb6cfe0d2ec271a2456660e54ba5e08',
    site: process?.env?.NEXT_PUBLIC_DATADOG_SITE || 'us5.datadoghq.com',
    forwardConsoleLogs: ['error', 'info'],
    forwardErrorsToLogs: true,
    sessionSampleRate: 100,
    env,
    service: 'dashboard',
    proxy: `/api/ddp`,
  })

  isDatadogInitialized = true
}

export const InitDatadogLogs: FC<{ env?: 'local' | 'stage' | 'prod' }> = ({
  env = 'local',
}) => {
  useEffect(() => {
    initDatadogLogs(env)
  }, [env])

  return null
}
