export default {
  title: 'Deploys/RenderedValues',
}

import { RenderedValues } from './RenderedValues'

export const ObjectFormat = () => (
  <RenderedValues
    values={{
      database_url: 'postgres://host:5432/mydb',
      redis_url: 'redis://cache:6379',
      api_key: 'sk-prod-abc123',
      region: 'us-east-1',
    }}
  />
)

export const ArrayFormat = () => (
  <RenderedValues
    values={[
      { name: 'database_url', value: 'postgres://host:5432/mydb' },
      { name: 'redis_url', value: 'redis://cache:6379' },
    ] as any}
  />
)
