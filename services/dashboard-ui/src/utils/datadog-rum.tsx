'use client'

import React, { type FC, useEffect } from 'react'
import { useUser } from '@auth0/nextjs-auth0/client'
import { datadogRum } from '@datadog/browser-rum'

export const InitDatadogRUM: FC<{ env?: "local" | "stage" | "prod" }> = ({ env = "local" }) => {
  const { user } = useUser() 
  
  useEffect(() => {
    const initDDLogs = () => {
      datadogRum.init({
        applicationId:
          process?.env?.NEXT_PUBLIC_DATADOG_APP_ID ||
          '19376b57-b3fb-4ad2-b0e9-fcdf9c986069',
        clientToken:
          process?.env?.NEXT_PUBLIC_DATADOG_CLIENT_TOKEN ||
          'pub6fb6cfe0d2ec271a2456660e54ba5e08',
        site: process?.env?.NEXT_PUBLIC_DATADOG_SITE || 'us5.datadoghq.com',
        env,
        service: 'dashboard',

        // collection settings
        sessionSampleRate: 100,
        sessionReplaySampleRate: 20,
        trackUserInteractions: true,
        trackResources: true,
        trackLongTasks: true,
        defaultPrivacyLevel: 'mask-user-input',
      })

      datadogRum.setUser({
        id: user?.sub,
        name: user?.name,
        email: user?.email,
        org_id: user?.org_id
      })
    }

    initDDLogs()
  }, [])

  return <></>
}
