'use client'

import { datadogRum } from '@datadog/browser-rum'
import React, { type FC, useEffect } from 'react'

export const InitDatadogRUM: FC = () => {
  useEffect(() => {
    const initDDLogs = () => {
      const ddEnv =
        process?.env?.NEXT_PUBLIC_DATADOG_ENV ||
        window.location?.host?.split('.')[1]

      datadogRum.init({
        applicationId:
          process?.env?.NEXT_PUBLIC_DATADOG_APP_ID ||
          '19376b57-b3fb-4ad2-b0e9-fcdf9c986069',
        clientToken:
          process?.env?.NEXT_PUBLIC_DATADOG_CLIENT_TOKEN ||
          'pub6fb6cfe0d2ec271a2456660e54ba5e08',
        site: process?.env?.NEXT_PUBLIC_DATADOG_SITE || 'us5.datadoghq.com',
        env: ddEnv || 'local',
        service: 'dashboard',

        // collection settings
        sessionSampleRate: 100,
        sessionReplaySampleRate: 20,
        trackUserInteractions: true,
        trackResources: true,
        trackLongTasks: true,
        defaultPrivacyLevel: 'mask-user-input',
      })
    }

    initDDLogs()
  }, [])

  return <></>
}
