import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { DebouncedSearchInput } from '@/components/common/DeboundedSearch'
import { TriggeredByFilter } from '@/components/actions/TriggeredByFilter'
import { LatestActionRunCard } from '@/components/actions/LatestActionRunCard'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { InstallActionRowActions } from '@/components/actions/InstallActionRowActions'
import { InstallCronOfflineBanner } from '@/components/installs/InstallCronOfflineBanner'
import { RunAdhocActionButton } from '@/components/installs/management/RunAdhocAction'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getInstallActionsLatestRuns } from '@/lib'
import type { TActionConfigTriggerType, TInstallAction } from '@/types'
import { isActionInAppConfig } from '@/utils/app-config-membership'
import {
  InstallActionsList,
  type TInstallActionListItem,
} from './InstallActionsList'

const LIMIT = 10

export const InstallActionsListContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  const { appConfig } = useInstallAppConfig()
  const [searchParams] = useSearchParams()

  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined
  const trigger_types = searchParams.get('trigger_types') || undefined

  const { data: result, isLoading } = useQuery({
    queryKey: [
      'install-actions',
      org?.id,
      install?.id,
      offset,
      q,
      trigger_types,
    ],
    queryFn: () =>
      getInstallActionsLatestRuns({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
        q,
        trigger_types,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: 20000,
    enabled: !!org?.id && !!install?.id,
  })

  const hasCronSchedule = !!appConfig?.action_workflow_configs?.some((config) =>
    config.triggers?.some((trigger) => trigger.type === 'cron')
  )

  const toListItem = (action: TInstallAction): TInstallActionListItem => {
    const actionId = action.action_workflow_id ?? action.id ?? ''
    const workflow = action.action_workflow
    const recentRun = action.runs?.[0]
    const config = workflow?.configs?.[0]
    const removed = !isActionInAppConfig(appConfig, actionId)
    const href = actionId
      ? installLink({
          installId: install?.id,
          appId: install?.app_id,
          suffix: `/actions/${actionId}`,
        })
      : undefined
    const runHref =
      recentRun?.id && actionId
        ? installLink({
            installId: install?.id,
            appId: install?.app_id,
            suffix: `/actions/${actionId}/runs/${recentRun.id}`,
          })
        : undefined

    return {
      id: actionId,
      name: workflow?.name ?? 'Action',
      href,
      removed,
      actions: (
        <InstallActionRowActions
          action={workflow}
          actionId={actionId}
          config={config}
          name={workflow?.name ?? 'Action'}
          removed={removed}
        />
      ),
      latestRun: (
        <LatestActionRunCard
          flush
          run={recentRun}
          href={runHref}
          trigger={
            recentRun ? (
              <ActionTriggerType
                componentName={recentRun.run_env_vars?.COMPONENT_NAME}
                componentPath={
                  recentRun.run_env_vars?.COMPONENT_ID
                    ? installLink({
                        installId: install?.id,
                        appId: install?.app_id,
                        suffix: `/components/${recentRun.run_env_vars.COMPONENT_ID}`,
                      })
                    : undefined
                }
                triggerType={
                  recentRun.triggered_by_type as TActionConfigTriggerType
                }
              />
            ) : null
          }
        />
      ),
    }
  }

  return (
    <InstallActionsList
      items={(result?.data ?? []).map(toListItem)}
      loading={isLoading}
      filtered={!!q || !!trigger_types}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
      banner={
        <InstallCronOfflineBanner
          runnerStatus={install?.runner_status}
          kind="action"
          hasCronSchedule={hasCronSchedule}
          orgId={org?.id}
          installId={install?.id}
        />
      }
      search={
        <DebouncedSearchInput
          className="w-full md:w-fit"
          labelClassName="w-full md:w-fit"
          placeholder="Search by name or ID..."
        />
      }
      filterActions={<TriggeredByFilter />}
      actions={<RunAdhocActionButton />}
    />
  )
}
