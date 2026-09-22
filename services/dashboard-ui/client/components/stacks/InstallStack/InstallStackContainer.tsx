import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { EditStackOverridesButton } from '@/components/installs/management/EditStackOverrides'
import { ReprovisionStackButton } from '@/components/installs/management/ReprovisionStack'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getInstallStack } from '@/lib'
import type { TAPIError, TCustomNestedStack } from '@/types'
import { InstallStack, type TInstallNestedStack } from './InstallStack'

const byCreatedAtDesc = (
  a: { created_at?: string },
  b: { created_at?: string }
) => {
  const aTime = a.created_at ? Date.parse(a.created_at) : 0
  const bTime = b.created_at ? Date.parse(b.created_at) : 0
  return bTime - aTime
}

const nestedStack = ({
  id,
  name,
  source,
  stack,
  type,
}: TInstallNestedStack) => ({
  id,
  name,
  source,
  stack,
  type,
})

const resolveNestedStacks = (
  appStacks: TCustomNestedStack[],
  installStacks: TCustomNestedStack[],
  runnerURL?: string,
  runnerOverrideURL?: string,
  vpcURL?: string,
  vpcOverrideURL?: string
) => {
  const stacks: TInstallNestedStack[] = []
  const effectiveRunnerURL = runnerOverrideURL || runnerURL
  const effectiveVpcURL = vpcOverrideURL || vpcURL

  if (effectiveRunnerURL) {
    stacks.push(
      nestedStack({
        id: 'runner',
        name: 'Runner',
        source: runnerOverrideURL ? 'install override' : 'app config',
        stack: { name: 'Runner', template_url: effectiveRunnerURL },
        type: 'runner',
      })
    )
  }

  if (effectiveVpcURL) {
    stacks.push(
      nestedStack({
        id: 'vpc',
        name: 'VPC',
        source: vpcOverrideURL ? 'install override' : 'app config',
        stack: { name: 'VPC', template_url: effectiveVpcURL },
        type: 'vpc',
      })
    )
  }

  const overrides = new Map(installStacks.map((stack) => [stack.name, stack]))

  for (const stack of appStacks) {
    const override = overrides.get(stack.name)
    stacks.push(
      nestedStack({
        id: `custom-${stack.name}`,
        name: stack.name ?? 'Custom stack',
        source: override ? 'install override' : 'app config',
        stack: override ?? stack,
        type: 'custom',
      })
    )
    overrides.delete(stack.name)
  }

  for (const stack of overrides.values()) {
    stacks.push(
      nestedStack({
        id: `custom-${stack.name}`,
        name: stack.name ?? 'Custom stack',
        source: 'install only',
        stack,
        type: 'custom',
      })
    )
  }

  return stacks
}

export const InstallStackContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const {
    appConfig,
    error: configError,
    isLoading: configLoading,
  } = useInstallAppConfig()

  const { data: stack, isLoading: versionsLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-stack', org?.id, install?.id],
    queryFn: () => getInstallStack({ orgId: org.id, installId: install.id }),
    refetchInterval: 20000,
    enabled: !!org?.id && !!install?.id,
  })

  const versions = [...(stack?.versions ?? [])].sort(byCreatedAtDesc)
  const appStack = appConfig?.stack
  const installConfig = install?.install_config
  const nestedStacks = resolveNestedStacks(
    appStack?.custom_nested_stacks ?? [],
    installConfig?.custom_nested_stacks ?? [],
    appStack?.runner_nested_template_url,
    installConfig?.runner_nested_template_url,
    appStack?.vpc_nested_template_url,
    installConfig?.vpc_nested_template_url
  )

  return (
    <InstallStack
      configVersion={appConfig?.version}
      stackName={appStack?.name}
      stackType={appStack?.type}
      nestedStacks={nestedStacks}
      configLoading={configLoading}
      configError={
        configError
          ? (configError as TAPIError).error || 'Unable to load stack config.'
          : undefined
      }
      configAction={<EditStackOverridesButton variant="secondary" />}
      versions={versions}
      versionsLoading={versionsLoading}
      latestVersionAction={
        versions.length ? (
          <ReprovisionStackButton size="sm" variant="secondary" />
        ) : null
      }
    />
  )
}
