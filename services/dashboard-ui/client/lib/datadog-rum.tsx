import { useParams } from 'react-router'
import { type FC, useEffect } from 'react'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { datadogRum } from '@datadog/browser-rum'

let isDatadogRUMInitialized = false

const initDatadogRUM = (env: 'local' | 'stage' | 'prod') => {
  if (isDatadogRUMInitialized) return

  datadogRum.init({
    applicationId: '19376b57-b3fb-4ad2-b0e9-fcdf9c986069',
    clientToken: 'pub6fb6cfe0d2ec271a2456660e54ba5e08',
    site: 'us5.datadoghq.com',
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
  const { orgId } = useParams<{ orgId: string }>()
  const { user } = useAuth()
  const { datadogEnv } = useConfig()

  useEffect(() => {
    initDatadogRUM(env)
  }, [env])

  useEffect(() => {
    if (isDatadogRUMInitialized && user) {
      datadogRum.setUser({
        id: user.sub,
        name: user.name,
        email: user.email,
        org_id: orgId,
      })
    }
  }, [user])

  return null
}
