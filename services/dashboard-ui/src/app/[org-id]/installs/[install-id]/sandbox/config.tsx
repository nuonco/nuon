// NOTE(nnnnat): needs updated to stratus

import { AppSandboxConfig, AppSandboxVariables, Notice } from '@/components'
import { OperationRolesList } from '@/components/common/OperationRolesList'
import { Text } from '@/components/common/Text'
import { ValuesFileModal } from '@/components/old/InstallSandbox'
import { getAppConfig } from '@/lib'

export const SandboxConfig = async ({
  appId,
  appConfigId,
  orgId,
}: {
  appId: string
  appConfigId: string
  orgId: string
}) => {
  const { data, error } = await getAppConfig({
    appConfigId,
    appId,
    orgId,
    recurse: true,
  })

  return error ? (
    <Notice>{error?.error}</Notice>
  ) : (
    <>
      <AppSandboxConfig sandboxConfig={data?.sandbox} />
      <AppSandboxVariables
        variables={data?.sandbox?.variables}
        isNotTruncated
      />
      {data?.sandbox?.operation_roles &&
        Object.keys(data.sandbox.operation_roles).length > 0 && (
          <div className="flex flex-col gap-4">
            <Text variant="body" weight="strong" level={5}>
              Operation Roles
            </Text>
            <OperationRolesList
              operationRoles={data.sandbox.operation_roles}
            />
          </div>
        )}
      <ValuesFileModal valuesFiles={data?.sandbox?.variables_files} />
    </>
  )
}
