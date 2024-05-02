'use client'

import { datadogRum } from '@datadog/browser-rum'
import React, { type FC, useEffect } from 'react'

export const InitDatadogRUM: FC = () => {
  useEffect(() => {
    const initDDLogs = () => {
      datadogRum.init({
        applicationId: process?.env?.NEXT_PUBLIC_DATADOG_APP_ID,
        clientToken: process?.env?.NEXT_PUBLIC_DATADOG_CLIENT_TOKEN,
        site: process?.env?.NEXT_PUBLIC_DATADOG_SITE,
        env: process?.env?.NEXT_PUBLIC_DATADOG_ENV,
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
