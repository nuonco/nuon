import { ComponentConfiguration, Text } from '@/components'
import { ValuesFileModal } from '@/components/old/InstallSandbox'
import { getAppConfigById } from '@/lib'
import type { TInstall } from '@/types'

export const ComponentConfig = async ({
  componentId,
  install,
  orgId,
}: {
  install: TInstall
  componentId: string
  orgId: string
}) => {
  const { data: config, error } = await getAppConfigById({
    appConfigId: install.app_config_id,
    appId: install.app_id,
    orgId,
    recurse: true,
  })

  const componentConfig = config?.component_config_connections?.find(
    (c) => c.component_id === componentId
  )

  return error ? (
    <Text>{error?.error}</Text>
  ) : componentConfig ? (
    <>
      <ComponentConfiguration config={componentConfig} isNotTruncated />
      {componentConfig?.terraform_module?.variables_files?.length ? (
        <ValuesFileModal
          valuesFiles={componentConfig?.terraform_module?.variables_files}
        />
      ) : null}
    </>
  ) : (
    <Text>No component config found.</Text>
  )
}
