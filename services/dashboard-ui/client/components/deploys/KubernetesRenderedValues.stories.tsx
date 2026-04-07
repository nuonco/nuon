import type { Meta, StoryObj } from '@ladle/react'
import { KubernetesRenderedValues } from './KubernetesRenderedValues'

export default {
  title: 'Deploys/KubernetesRenderedValues',
} satisfies Meta

export const Default: StoryObj = {
  render: () => (
    <KubernetesRenderedValues
      values={[
        { name: 'DATABASE_URL', value: 'postgres://host:5432/mydb' },
        { name: 'REDIS_URL', value: 'redis://cache:6379' },
        { name: 'LOG_LEVEL', value: 'info' },
      ] as any}
    />
  ),
}

export const Empty: StoryObj = {
  render: () => <KubernetesRenderedValues values={[]} />,
}
