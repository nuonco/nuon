export default {
  title: 'Apps/Config/AppRunner',
}

import { AppRunner } from './AppRunner'

export const Default = () => (
  <AppRunner
    appConfig={{
      runner: {
        cloud_platform: 'aws',
        app_runner_type: 'eks',
        helm_driver: 'secrets',
        env_vars: {
          LOG_LEVEL: 'info',
          MAX_CONNECTIONS: '100',
        },
      },
    } as any}
  />
)

export const Minimal = () => (
  <AppRunner
    appConfig={{
      runner: {
        cloud_platform: 'aws',
        app_runner_type: 'eks',
        env_vars: {},
      },
    } as any}
  />
)

export const WithInitScript = () => (
  <AppRunner
    appConfig={{
      runner: {
        cloud_platform: 'aws',
        app_runner_type: 'eks',
        init_script: 'https://example.com/init.sh',
        env_vars: {},
      },
    } as any}
  />
)
