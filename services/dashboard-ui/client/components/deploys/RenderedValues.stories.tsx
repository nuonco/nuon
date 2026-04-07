import type { Meta, StoryObj } from '@ladle/react'
import { RenderedValues } from './RenderedValues'

export default {
  title: 'Deploys/RenderedValues',
} satisfies Meta

export const ObjectFormat: StoryObj = {
  render: () => (
    <RenderedValues
      values={{
        database_url: 'postgres://host:5432/mydb',
        redis_url: 'redis://cache:6379',
        api_key: 'sk-prod-abc123',
        region: 'us-east-1',
      }}
    />
  ),
}

export const ArrayFormat: StoryObj = {
  render: () => (
    <RenderedValues
      values={[
        { name: 'database_url', value: 'postgres://host:5432/mydb' },
        { name: 'redis_url', value: 'redis://cache:6379' },
      ] as any}
    />
  ),
}
