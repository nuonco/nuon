import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { pickCreateInstallConfig } from '@/components/installs/CreateInstall/CreateInstallFromAppContainer'
import { FormSkeleton } from '@/components/installs/CreateInstall/FormSkeleton'
import { normalizeInstallPlatform } from '@/components/installs/forms/InstallForm'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useOrg } from '@/hooks/use-org'
import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'
import {
  getAppConfig,
  getActions,
  getBranchConfigs,
  getCloudConnections,
  getComponents,
} from '@/lib'
import { latestBranchConfig } from '@/utils/branch-utils'
import type { TCloudConnectionSummary } from '@/types'
import { CreateInstallWizardSession } from './CreateInstallWizardSession'

const errorMessage = (err: unknown) =>
  (err as { error?: string } | null)?.error || 'Unable to load app configuration'

interface ICreateInstallWizardContainer {
  installsPath: string
}

export const CreateInstallWizardContainer = ({
  installsPath,
}: ICreateInstallWizardContainer) => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const requireTargetAccount = useOrgFeatureFlag('phone-home-auth')
  const platform = normalizeInstallPlatform(app.runner_config?.app_runner_type)
  const cloudConnectionsEnabled = platform === 'aws' || platform === 'azure'
  const groupConfig = latestBranchConfig(branch)

  const {
    data: branchConfigs,
    isLoading: branchConfigsLoading,
    error: branchConfigsError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-branch-app-configs', org.id, app.id, branch.id],
    queryFn: () =>
      getBranchConfigs({
        orgId: org.id,
        appId: app.id,
        branchId: branch.id,
        limit: 100,
      }),
    enabled: !!org.id && !!app.id && !!branch.id,
  })

  const configId = pickCreateInstallConfig(branchConfigs, {
    requireActive: true,
    fallbackToFirst: false,
  })?.id

  const {
    data: config,
    isLoading: configLoading,
    error: configError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', org.id, app.id, configId],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: app.id,
        appConfigId: configId!,
        recurse: true,
      }),
    enabled: !!org.id && !!configId,
  })

  const { data: cloudConnections, isLoading: cloudConnectionsLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['cloud-connections', org.id, 'all'],
    queryFn: async () => {
      const all: TCloudConnectionSummary[] = []
      let offset = 0
      for (;;) {
        const { data, pagination } = await getCloudConnections({
          orgId: org.id,
          offset,
          limit: 100,
        })
        const page = data ?? []
        all.push(...page)
        if (!pagination?.hasNext || page.length === 0) return all
        offset += page.length
      }
    },
    enabled: !!org.id && cloudConnectionsEnabled,
  })

  const componentIds = config?.component_ids ?? []
  const actionIds = config?.action_ids ?? []
  const needsComponents = componentIds.length > 0
  const { data: componentsResult, isLoading: componentsLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['components', org.id, app.id, 'create-install-gate'],
    queryFn: () => getComponents({ orgId: org.id, appId: app.id, limit: 100 }),
    enabled: !!org.id && !!app.id && needsComponents && !!config,
  })

  const { data: actionsResult, isLoading: actionsLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['actions', org.id, app.id, branch.id, 'create-install-review'],
    queryFn: () =>
      getActions({
        orgId: org.id,
        appId: app.id,
        branch_id: branch.id,
        limit: 100,
      }),
    enabled: !!org.id && !!app.id && actionIds.length > 0 && !!config,
  })

  const isLoading =
    branchConfigsLoading ||
    configLoading ||
    (needsComponents && componentsLoading) ||
    (actionIds.length > 0 && actionsLoading) ||
    (cloudConnectionsEnabled && cloudConnectionsLoading)

  const missingConfig =
    !branchConfigsLoading && !branchConfigsError && !configId
      ? { error: 'This branch has no active app config.' }
      : undefined
  const loadError = branchConfigsError || configError || missingConfig

  if (!app.runner_config) {
    return (
      <div className="mx-auto flex w-full max-w-4xl flex-col gap-4 px-6 pt-8">
        <Banner theme="error">This app cannot create installs yet.</Banner>
        <Button variant="secondary" href={installsPath} className="w-fit">
          Back to installs
        </Button>
      </div>
    )
  }

  if (loadError && !isLoading) {
    return (
      <div className="mx-auto flex w-full max-w-4xl flex-col gap-4 px-6 pt-8">
        <Banner theme="error">{errorMessage(loadError)}</Banner>
        <Button variant="secondary" href={installsPath} className="w-fit">
          Back to installs
        </Button>
      </div>
    )
  }

  const inputConfig = config?.input
    ? {
        ...config.input,
        input_groups: (config.input.input_groups || []).map((group) => ({
          ...group,
          app_inputs:
            config.input?.inputs?.filter((input) => input.group_id === group.id) ||
            [],
        })),
      }
    : undefined

  if (isLoading) {
    return (
      <div className="mx-auto w-full max-w-4xl px-6 pt-8">
        <FormSkeleton />
      </div>
    )
  }

  if (!inputConfig) {
    return (
      <div className="mx-auto flex w-full max-w-4xl flex-col gap-4 px-6 pt-8">
        <Banner theme="error">This branch has no active app config.</Banner>
        <Button variant="secondary" href={installsPath} className="w-fit">
          Back to installs
        </Button>
      </div>
    )
  }

  return (
    <CreateInstallWizardSession
      app={app}
      branch={branch}
      groupConfig={groupConfig}
      inputConfig={inputConfig}
      actions={(actionsResult?.data ?? []).filter((action) =>
        actionIds.includes(action.id)
      )}
      components={(componentsResult?.data ?? []).filter((component) =>
        componentIds.includes(component.id)
      )}
      requireTargetAccount={requireTargetAccount}
      cloudConnections={
        platform === 'aws'
          ? (cloudConnections || []).filter(
              (connection) =>
                connection.platform === 'aws' && connection.status === 'verified'
            )
          : undefined
      }
      installsPath={installsPath}
    />
  )
}
