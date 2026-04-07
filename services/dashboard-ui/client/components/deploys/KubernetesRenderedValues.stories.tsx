export default {
  title: 'Deploys/KubernetesRenderedValues',
}

import { KubernetesRenderedValues } from './KubernetesRenderedValues'

export const Default = () => (
  <KubernetesRenderedValues
    values={[
      { name: 'DATABASE_URL', value: 'postgres://host:5432/mydb' },
      { name: 'REDIS_URL', value: 'redis://cache:6379' },
      { name: 'LOG_LEVEL', value: 'info' },
    ] as any}
  />
)

export const Empty = () => <KubernetesRenderedValues values={[]} />
