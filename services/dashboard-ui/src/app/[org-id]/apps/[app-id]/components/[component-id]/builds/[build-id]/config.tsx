import { ComponentConfiguration, Text } from '@/components'

import type { TComponentConfig } from '@/types'
import { nueQueryData } from '@/utils'

export const ComponentConfig = async ({
  componentId,
  componentConfigId,
  orgId,
}: {
  componentId: string
  componentConfigId: string
  orgId: string
}) => {
  const { data: componentConfig, error } = await nueQueryData<TComponentConfig>(
    {
      orgId,
      path: `components/${componentId}/configs/${componentConfigId}`,
    }
  )
  return error ? (
    <Text>{error?.error}</Text>
  ) : (
    <ComponentConfiguration config={componentConfig} />
  )
}
