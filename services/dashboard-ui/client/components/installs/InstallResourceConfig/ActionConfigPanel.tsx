import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { ActionStep } from '@/components/actions/ActionStep'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { getActionConfig } from '@/lib'
import type { TActionConfig, TActionConfigTriggerType } from '@/types'
import { sortByIdx } from '@/utils/action-utils'
import {
  InstallConfigSource,
  type IInstallConfigSource,
} from './InstallConfigSource'
import { ConfigProperties, ConfigSection } from './ConfigLayout'
import { useInstallConfigRun } from './use-install-config-run'

export interface IActionConfigDetails extends IInstallConfigSource {
  config?: TActionConfig
  isLoadingSteps?: boolean
}

export const ActionConfigDetails = ({
  config,
  isLoadingSteps,
  ...source
}: IActionConfigDetails) => {
  const { install } = useInstall()
  const installLink = useInstallLink()
  const resourceHref = (suffix: string) =>
    installLink({
      orgId: install?.org_id,
      appId: install?.app_id,
      installId: install?.id,
      suffix,
    })
  const steps = sortByIdx(config?.steps ?? [])
  const kubeEnabled = !!config?.enable_kube_config?.bool

  if (!config) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No configuration"
        emptyMessage="This action is not in the app config this install is using."
      />
    )
  }

  return (
    <>
      <InstallConfigSource {...source} />
      <ConfigSection title="Configuration">
        <ConfigProperties
          properties={[
            {
              label: 'Kube config',
              value: (
                <Badge size="sm" theme={kubeEnabled ? 'info' : 'neutral'}>
                  {kubeEnabled ? 'Enabled' : 'Disabled'}
                </Badge>
              ),
            },
            {
              label: 'Timeout',
              value: config.timeout ? (
                <Duration nanoseconds={config.timeout} variant="subtext" />
              ) : undefined,
            },
            { label: 'Container image', value: config.image },
            { label: 'Execution role', value: config.role },
            {
              label: 'Kubernetes context',
              value: config.kubernetes_context_name,
            },
          ]}
        />
      </ConfigSection>
      {config.triggers?.length ? (
        <ConfigSection title="Triggers">
          <div className="flex flex-col gap-3 rounded-md border px-4 py-3">
            {config.triggers.map((trigger) => (
              <ActionTriggerType
                key={trigger.id}
                componentName={trigger.component?.name}
                componentPath={
                  trigger.component_id
                    ? resourceHref(`/components/${trigger.component_id}`)
                    : undefined
                }
                cronSchedule={trigger.cron_schedule}
                triggerType={trigger.type as TActionConfigTriggerType}
              />
            ))}
          </div>
        </ConfigSection>
      ) : null}
      <ConfigSection title="Steps">
        {isLoadingSteps && !steps.length ? (
          <Loading />
        ) : steps.length ? (
          steps.map((step, index) => (
            <ActionStep key={step.id ?? index} index={index} step={step} />
          ))
        ) : (
          <Text variant="subtext" theme="neutral">
            This action has no steps configured.
          </Text>
        )}
      </ConfigSection>
    </>
  )
}

const ActionConfigPanelBody = ({ config }: { config?: TActionConfig }) => {
  const { org } = useOrg()
  const { isLoading, run, runHref } = useInstallConfigRun(config?.app_config_id)
  const { data: fullConfig, isLoading: isLoadingSteps } = useQuery({
    queryKey: ['action-config', org?.id, config?.id],
    queryFn: () =>
      getActionConfig({ orgId: org!.id, actionConfigId: config!.id! }),
    enabled: !!org?.id && !!config?.id,
  })

  return (
    <ActionConfigDetails
      config={
        config
          ? { ...config, steps: fullConfig?.steps ?? config.steps }
          : undefined
      }
      isLoading={isLoading}
      isLoadingSteps={isLoadingSteps}
      run={run}
      runHref={runHref}
    />
  )
}

export const ActionConfigPanel = ({
  config,
  name,
  ...props
}: {
  config?: TActionConfig
  name: string
} & Omit<IPanel, 'heading' | 'children'>) => (
  <Panel
    heading={`${name} config`}
    panelKey={`action-config-${config?.action_workflow_id ?? name}`}
    size="half"
    {...props}
  >
    <ActionConfigPanelBody config={config} />
  </Panel>
)
