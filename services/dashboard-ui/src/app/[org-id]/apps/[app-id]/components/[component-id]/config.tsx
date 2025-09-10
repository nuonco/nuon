import { ComponentConfiguration, Text } from '@/components'
import type { TComponentConfig } from '@/types'
import { api } from '@/lib/api'

// TODO(nnnat): get the component config form the app config
export const Config = async ({
  componentId,
  orgId,
}: {
  componentId: string
  orgId: string
}) => {
  const { data: componentConfig, error } = await api<TComponentConfig>({
    orgId,
    path: `components/${componentId}/configs/latest`,
  })

  return error ? (
    <Text>{error?.error}</Text>
  ) : (
    <ComponentConfiguration config={componentConfig} isNotTruncated />
  )
}
