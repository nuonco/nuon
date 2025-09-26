import { AppSandboxConfig, AppSandboxVariables, Notice } from '@/components'
import { ValuesFileModal } from '@/components/InstallSandbox'
import { getAppConfigById } from '@/lib'

export const SandboxConfig = async ({
  appId,
  appConfigId,
  orgId,
}: {
  appId: string
  appConfigId: string
  orgId: string
}) => {
  const { data, error } = await getAppConfigById({
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
      <ValuesFileModal valuesFiles={data?.sandbox?.variables_files} />
    </>
  )
}
