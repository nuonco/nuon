import { ComponentConfiguration, Text } from '@/components'
import { getAppConfigById } from '@/lib'

export const ComponentConfig = async ({
  appConfigId,
  appId,
  componentId,
  orgId,
}: {
  appConfigId: string
  appId: string
  componentId: string
  orgId: string
}) => {
  const { data: config, error } = await getAppConfigById({
    appConfigId,
    appId,
    orgId,
    recurse: true,
  })

  const componentConfig = config?.component_config_connections?.find(
    (c) => c.component_id === componentId
  )

  return error ? (
    <Text>{error?.error}</Text>
  ) : componentConfig ? (
    <ComponentConfiguration config={componentConfig} hideHelmValuesFile />
  ) : (
    <Text>No component config found.</Text>
  )
}
