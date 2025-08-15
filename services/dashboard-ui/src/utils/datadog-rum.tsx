'use client'

import { type FC, useEffect } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { datadogRum } from '@datadog/browser-rum'

let isDatadogRUMInitialized = false

const initDatadogRUM = (env: 'local' | 'stage' | 'prod') => {
  if (isDatadogRUMInitialized) return

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
    proxy: `/api/ddp`,
  })

  isDatadogRUMInitialized = true
}

export const InitDatadogRUM: FC<{ env?: 'local' | 'stage' | 'prod' }> = ({
  env = 'local',
}) => {
  const { user } = useUser()

  useEffect(() => {
    initDatadogRUM(env)
  }, [env])

  useEffect(() => {
    if (isDatadogRUMInitialized && user) {
      datadogRum.setUser({
        id: user.sub,
        name: user.name,
        email: user.email,
        org_id: user.org_id,
      })
    }
  }, [user])

  return null
}
