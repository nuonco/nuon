import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { InstallConfigsTimelineComponent } from '@/components/install-configs/InstallConfigsTimeline'
import { InstallVersionsTimeline } from '@/components/install-versions/InstallVersionsTimeline'
import { EditInputsButton } from '@/components/installs/management/EditInputs'
import { Toast } from '@/components/surfaces/Toast'
import { Text } from '@/components/common/Text'
import { useCurrentAppBranchRun } from '@/hooks/use-current-app-branch-run'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import {
  generateCLIInstallConfig,
  getAppBranch,
  getInstallConfigVersions,
  getInstallCurrentInputs,
  syncInstallConfig,
} from '@/lib'
import { normalizeAppInputGroups } from '@/utils/app-utils'
import { COMPONENT_OVERRIDE_INPUT_GROUP } from '@/utils/install-utils'
import {
  ConfigSyncAction,
  InstallConfigurationAppBranch,
  InstallConfigurationConfigFile,
  InstallConfigurationInputs,
  InstallConfigurationOverrides,
} from './InstallConfiguration'

const branchRunHref = ({
  appId,
  branchId,
  orgId,
  runId,
}: {
  appId?: string
  branchId?: string
  orgId?: string
  runId?: string
}) =>
  appId && branchId && orgId && runId
    ? `/${orgId}/apps/${appId}/branches/${branchId}/runs/${runId}`
    : undefined

export const InstallConfigurationAppBranchContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { run: appliedRun, isLoading: appliedRunLoading } =
    useCurrentAppBranchRun()
  const branchId = install?.app_branch?.id

  const { data: branch, isLoading: branchLoading } = useQuery({
    queryKey: [
      'install-configuration-app-branch',
      org?.id,
      install?.app_id,
      branchId,
    ],
    queryFn: () =>
      getAppBranch({
        orgId: org!.id,
        appId: install!.app_id!,
        branchId: branchId!,
        latestConfig: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!branchId,
  })

  const latestRun = branch?.latest_run
  const latestConfig = latestRun?.app_branch_config ?? branch?.configs?.at(0)
  const branchHref =
    org?.id && install?.app_id && branchId
      ? `/${org.id}/apps/${install.app_id}/branches/${branchId}`
      : undefined

  return (
    <InstallConfigurationAppBranch
      appliedConfigId={install?.app_config_id}
      branchName={install?.app_branch?.name}
      branchHref={branchHref}
      branchConfig={latestConfig}
      latestRun={latestRun}
      latestRunHref={branchRunHref({
        orgId: org?.id,
        appId: install?.app_id,
        branchId,
        runId: latestRun?.id,
      })}
      appliedRun={appliedRun}
      appliedRunHref={branchRunHref({
        orgId: org?.id,
        appId: install?.app_id,
        branchId: appliedRun?.app_branch?.id ?? branchId,
        runId: appliedRun?.id,
      })}
      isLoading={branchLoading || appliedRunLoading}
      history={<InstallVersionsTimeline />}
    />
  )
}

const useInstallConfigurationInputs = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { appConfig, isLoading: configLoading } = useInstallAppConfig()

  const { data: inputs, isLoading: inputsLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-inputs', org?.id, install?.id],
    queryFn: () =>
      getInstallCurrentInputs({
        orgId: org!.id,
        installId: install!.id,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const groups = appConfig
    ? normalizeAppInputGroups(
        appConfig.input?.input_groups ?? [],
        appConfig.input?.inputs ?? []
      )
    : []

  return {
    groups,
    isLoading: configLoading || inputsLoading,
    values: inputs?.redacted_values ?? {},
  }
}

export const InstallConfigurationInputsContainer = () => {
  const { groups, isLoading, values } = useInstallConfigurationInputs()
  const inputGroups = groups.filter(
    (group) => group.name !== COMPONENT_OVERRIDE_INPUT_GROUP
  )

  return (
    <InstallConfigurationInputs
      action={<EditInputsButton variant="secondary" />}
      groups={inputGroups}
      isLoading={isLoading}
      values={values}
    />
  )
}

export const InstallConfigurationOverridesContainer = () => {
  const { groups, isLoading, values } = useInstallConfigurationInputs()
  const overrideGroup = groups.find(
    (group) => group.name === COMPONENT_OVERRIDE_INPUT_GROUP
  )

  return (
    <InstallConfigurationOverrides
      action={<EditInputsButton variant="secondary" />}
      inputs={overrideGroup?.app_inputs}
      isLoading={isLoading}
      values={values}
    />
  )
}

export const InstallConfigurationConfigFileContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const isManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  const { data: generatedConfig, isLoading: configLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-generate-cli-config', org?.id, install?.id],
    queryFn: () =>
      generateCLIInstallConfig({
        orgId: org!.id,
        installId: install!.id,
      }),
    enabled: !!org?.id && !!install?.id && isManagedByConfig,
  })

  const { data: versions, isLoading: versionsLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-config-versions', org?.id, install?.id],
    queryFn: () =>
      getInstallConfigVersions({
        orgId: org!.id,
        installId: install!.id,
      }),
    enabled: !!org?.id && !!install?.id && isManagedByConfig,
  })

  const { mutate: syncNow, isPending } = useMutation({
    mutationFn: () =>
      syncInstallConfig({
        orgId: org!.id,
        installId: install!.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['install-config-versions', org?.id, install?.id],
      })
      addToast(
        <Toast heading="Syncing config" theme="info">
          <Text>Syncing configs for {install?.name}.</Text>
        </Toast>
      )
    },
    onError: (error) => {
      addToast(
        <Toast heading="Config sync failed" theme="error">
          <Text>{error?.error || 'Unable to trigger config sync.'}</Text>
        </Toast>
      )
    },
  })

  const latestVersion = versions?.at(0)

  return (
    <InstallConfigurationConfigFile
      action={
        <ConfigSyncAction isPending={isPending} onSync={() => syncNow()} />
      }
      content={generatedConfig?.content}
      filename={latestVersion?.file_path ?? generatedConfig?.filename}
      history={
        <InstallConfigsTimelineComponent
          versions={versions ?? []}
          isLoading={versionsLoading}
          orgId={org?.id}
          installId={install?.id}
        />
      }
      isLoading={configLoading}
      isManagedByConfig={isManagedByConfig}
      latestVersionId={latestVersion?.id}
      syncedAt={latestVersion?.created_at}
    />
  )
}
