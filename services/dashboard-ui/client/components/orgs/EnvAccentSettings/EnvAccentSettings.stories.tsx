export default {
  title: 'Orgs/EnvAccentSettings',
}

import { EnvAccentSettings } from './EnvAccentSettings'

export const Default = () => (
  <EnvAccentSettings
    config={{
      label_key: 'env',
      values: {
        production: 'error',
        staging: 'warn',
        dev: 'success',
        qa: 'info',
      },
    }}
    isPending={false}
    error={null}
    onSubmit={(next) => {
      // eslint-disable-next-line no-console
      console.log('submit', next)
    }}
  />
)

export const Empty = () => (
  <EnvAccentSettings
    config={{ label_key: 'env', values: {} }}
    isPending={false}
    error={null}
    onSubmit={(next) => {
      // eslint-disable-next-line no-console
      console.log('submit', next)
    }}
  />
)

export const Saving = () => (
  <EnvAccentSettings
    config={{
      label_key: 'env',
      values: { production: 'error', staging: 'warn' },
    }}
    isPending
    error={null}
    onSubmit={() => undefined}
  />
)

export const WithError = () => (
  <EnvAccentSettings
    config={{ label_key: 'env', values: { production: 'error' } }}
    isPending={false}
    error={{ error: 'invalid accent color "purple" for value "production"' }}
    onSubmit={() => undefined}
  />
)
