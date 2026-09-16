import { useMemo, useState } from 'react'
import { DateTime } from 'luxon'
import {
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useNavigate, useSearchParams } from 'react-router'
import {
  createAppInstall,
  getAppBranch,
  getAppBranches,
  getAppConfig,
  getAppConfigs,
  getApps,
  getBranchConfigs,
  getComponents,
  getInstall,
} from '@/lib'
import type { TApp, TAppBranch, TComponent } from '@/types'
import { useToast } from '../../../hooks/use-toast'
import { useOrg } from '../../../providers/org-provider'
import { appSourceFromApp } from '../../molecules/AppSource'
import type { IAppSelectItem } from '../AppSelect'
import { installSetupHref } from '../../../utils/hrefs'
import {
  buildCreateInstallBody,
  installSetupInputs,
  normalizeInstallPlatform,
  resolveInstallConfig,
  type ICreateInstallValues,
} from '../../../utils/create-install'
import {
  InstallSetup,
  type IInstallSetupBranch,
  type IInstallSetupGroup,
} from './InstallSetup'

const appPlatform = (app?: TApp) =>
  normalizeInstallPlatform(
    app?.runner_config?.app_runner_type ??
      app?.runner_config?.cloud_platform ??
      app?.cloud_platform
  )

const hasActiveBuild = (components?: TComponent[]) =>
  (components ?? []).some(
    (component) => component?.latest_build?.status_v2?.status === 'active'
  )

const appConfigForReadiness = (app: TApp) =>
  app?.app_configs?.find(
    (config) =>
      (!config?.status || config.status === 'active') &&
      config?.labels?.source !== 'git-preview-run'
  )

const groupFromApi = (
  group: NonNullable<
    NonNullable<TAppBranch['configs']>[number]['install_groups']
  >[number]
): IInstallSetupGroup | undefined => {
  if (!group?.id) return undefined
  const labels = group?.label_selector?.match_labels ?? {}
  const values = Object.values(labels)
  const kind = group?.all_installs
    ? 'all'
    : Object.keys(labels).length === 0
      ? 'install-ids'
      : values.some((value) => value === '*')
        ? 'wildcard'
        : 'labels'

  return {
    id: group.id,
    name: group?.name ?? group.id,
    kind,
    labels,
  }
}

const branchFromApi = (branch: TAppBranch): IInstallSetupBranch | undefined => {
  if (!branch?.id) return undefined
  return {
    id: branch.id,
    name: branch?.name ?? branch.id,
    groups: (branch?.configs?.at(0)?.install_groups ?? []).flatMap((group) => {
      const resolved = groupFromApi(group)
      return resolved ? [resolved] : []
    }),
  }
}

export const InstallSetupContainer = () => {
  const { org, orgId } = useOrg()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addToast } = useToast()
  const [params] = useSearchParams()
  const installId = params.get('installId') || undefined
  const [selectedAppId, setSelectedAppId] = useState('')
  const [selectedBranchId, setSelectedBranchId] = useState<string>()
  const persistence = useMemo(
    () => (orgId ? { orgId, wizard: 'install-setup' } : undefined),
    [orgId]
  )

  const { data: appResult, isLoading: appsLoading } = useQuery({
    queryKey: ['apps', orgId, 'install-setup'],
    queryFn: () => getApps({ orgId: orgId!, limit: 100, offset: 0 }),
    enabled: !!orgId && !installId,
    staleTime: 30_000,
  })
  const rawApps = appResult?.data ?? []

  const appsWithComponents = rawApps.filter((app) => {
    const config = appConfigForReadiness(app)
    return !!config?.component_ids?.length
  })
  const componentQueries = useQueries({
    queries: appsWithComponents.map((app) => ({
      queryKey: ['components', orgId, app?.id, 'install-setup'],
      queryFn: () =>
        getComponents({
          orgId: orgId!,
          appId: app.id!,
          limit: 100,
          offset: 0,
        }),
      enabled: !!orgId && !!app?.id,
      staleTime: 30_000,
    })),
  })
  const componentsByApp = new Map(
    appsWithComponents.flatMap((app, index) =>
      app?.id ? [[app.id, componentQueries[index]?.data?.data] as const] : []
    )
  )

  const apps = rawApps.flatMap((app) => {
    if (!app?.id) return []
    const platform = appPlatform(app)
    const config = appConfigForReadiness(app)
    const components = componentsByApp.get(app.id)
    const readiness: IAppSelectItem['readiness'] =
      !app?.runner_config?.app_runner_type || platform === 'unknown'
        ? 'not-provisionable'
        : !config
          ? 'no-valid-config'
          : !config?.component_ids?.length
            ? 'no-components'
            : components && !hasActiveBuild(components)
              ? 'no-component-builds'
              : undefined
    const updated = app?.updated_at
      ? DateTime.fromISO(app.updated_at).toRelative()
      : undefined

    return [
      {
        id: app.id,
        name: app?.name ?? 'Unnamed app',
        platform,
        source: appSourceFromApp(app),
        updatedLabel: updated ? `synced ${updated}` : 'never synced',
        readiness,
      },
    ]
  })

  const { data: branchResult } = useQuery({
    queryKey: ['app-branches', orgId, selectedAppId, 'install-setup'],
    queryFn: () =>
      getAppBranches({
        orgId: orgId!,
        appId: selectedAppId,
        limit: 100,
        offset: 0,
      }),
    enabled: !!orgId && !!selectedAppId && !installId,
  })
  const rawBranches = branchResult?.data ?? []
  const branchQueries = useQueries({
    queries: rawBranches.map((branch) => ({
      queryKey: ['app-branch-with-config', orgId, selectedAppId, branch?.id],
      queryFn: () =>
        getAppBranch({
          orgId: orgId!,
          appId: selectedAppId,
          branchId: branch.id!,
          latestConfig: true,
        }),
      enabled: !!orgId && !!selectedAppId && !!branch?.id && !installId,
      staleTime: 30_000,
    })),
  })
  const branches = rawBranches.flatMap((branch, index) => {
    const resolved = branchFromApi(branchQueries[index]?.data ?? branch)
    return resolved ? [resolved] : []
  })

  const { data: configList, isLoading: configsLoading } = useQuery({
    queryKey: [
      selectedBranchId ? 'app-branch-app-configs' : 'app-configs',
      orgId,
      selectedAppId,
      selectedBranchId,
      'install-setup',
    ],
    queryFn: () =>
      selectedBranchId
        ? getBranchConfigs({
            orgId: orgId!,
            appId: selectedAppId,
            branchId: selectedBranchId,
            limit: 100,
          })
        : getAppConfigs({
            orgId: orgId!,
            appId: selectedAppId,
            limit: 100,
            offset: 0,
          }),
    enabled: !!orgId && !!selectedAppId && !installId,
  })
  const { config: selectedConfig, branchId: effectiveBranchId } =
    resolveInstallConfig(configList, selectedBranchId)
  const { data: config, isLoading: configLoading } = useQuery({
    queryKey: [
      'app-config',
      orgId,
      selectedAppId,
      selectedConfig?.id,
      'install-setup',
    ],
    queryFn: () =>
      getAppConfig({
        orgId: orgId!,
        appId: selectedAppId,
        appConfigId: selectedConfig!.id!,
        recurse: true,
      }),
    enabled: !!orgId && !!selectedAppId && !!selectedConfig?.id && !installId,
  })

  const inputs = useMemo(() => installSetupInputs(config?.input), [config])

  const { data: existingInstall } = useQuery({
    queryKey: ['install', orgId, installId],
    queryFn: () => getInstall({ orgId: orgId!, installId: installId! }),
    enabled: !!orgId && !!installId,
    refetchInterval: 4_000,
  })
  const platform = installId
    ? normalizeInstallPlatform(existingInstall?.cloud_platform)
    : appPlatform(rawApps.find((app) => app?.id === selectedAppId))

  const mutation = useMutation({
    mutationFn: (values: ICreateInstallValues) => {
      const selectedGroup = branches
        .find((branch) => branch.id === values.branchId)
        ?.groups.find((group) => group.id === values.installGroupId)
      return createAppInstall({
        orgId: orgId!,
        appId: selectedAppId,
        body: buildCreateInstallBody({
          values: {
            ...values,
            branchId: values.branchId || effectiveBranchId,
          },
          platform,
          groupLabels: selectedGroup?.labels,
        }),
      })
    },
    onSuccess: (result) => {
      const created = result.data
      void queryClient.invalidateQueries({ queryKey: ['installs', orgId] })
      queryClient.setQueryData(['install', orgId, created?.id], created)
      addToast({
        heading: 'Install created',
        description: `Created ${created?.name ?? 'the install'}. Provisioning may take a few minutes.`,
        theme: 'success',
      })
      if (created?.id && orgId) {
        navigate(installSetupHref(orgId, created.id), { replace: true })
      }
    },
  })

  return (
    <InstallSetup
      apps={apps}
      appsLoading={appsLoading}
      branches={branches}
      inputs={inputs}
      platform={platform}
      configurationKey={config?.id}
      configurationLoading={configsLoading || configLoading}
      configurationReady={!!config}
      requireTargetAccount={!!org?.features?.['phone-home-auth']}
      pending={mutation.isPending}
      error={mutation.error}
      install={existingInstall}
      persistence={persistence}
      onAppChange={setSelectedAppId}
      onBranchChange={setSelectedBranchId}
      onClearError={mutation.reset}
      onDiscard={() => navigate(`/${orgId}/installs`)}
      onSubmit={(values) => mutation.mutate(values)}
    />
  )
}
