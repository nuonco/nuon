import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { RunbookStep } from '@/components/runbooks/RunbookStep'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import type { TRunbookConfig } from '@/lib/ctl-api/apps/runbooks'
import { sortByIdx } from '@/utils/action-utils'
import { ConfigProperties, ConfigSection } from './ConfigLayout'
import {
  InstallConfigSource,
  type IInstallConfigSource,
} from './InstallConfigSource'
import { useInstallConfigRun } from './use-install-config-run'

export interface IRunbookConfigDetails extends IInstallConfigSource {
  config?: TRunbookConfig
}

export const RunbookConfigDetails = ({
  config,
  ...source
}: IRunbookConfigDetails) => {
  const { install } = useInstall()
  const installLink = useInstallLink()
  const steps = sortByIdx(config?.steps ?? [])
  const inputs = sortByIdx(config?.inputs ?? [])

  if (!config) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No configuration"
        emptyMessage="This runbook is not in the app config this install is using."
      />
    )
  }

  return (
    <>
      <InstallConfigSource {...source} />
      {inputs.length ? (
        <ConfigSection title="Inputs">
          <ConfigProperties
            properties={inputs.map((input) => ({
              label: input.display_name || input.name || '',
              value: (
                <span className="flex flex-wrap items-center gap-2">
                  <Text variant="subtext" family="mono" className="break-all">
                    {input.name}
                  </Text>
                  {input.required ? (
                    <Badge size="sm" theme="neutral">
                      Required
                    </Badge>
                  ) : null}
                </span>
              ),
            }))}
          />
        </ConfigSection>
      ) : null}
      <ConfigSection title="Steps">
        {steps.length ? (
          steps.map((step, index) => (
            <RunbookStep
              key={step.id ?? index}
              actionBasePath={installLink({
                orgId: install?.org_id,
                appId: install?.app_id,
                installId: install?.id,
              })}
              index={index}
              step={step}
            />
          ))
        ) : (
          <Text variant="subtext" theme="neutral">
            This runbook has no steps configured.
          </Text>
        )}
      </ConfigSection>
    </>
  )
}

const RunbookConfigPanelBody = ({ config }: { config?: TRunbookConfig }) => {
  const { isLoading, run, runHref } = useInstallConfigRun(config?.app_config_id)

  return (
    <RunbookConfigDetails
      config={config}
      isLoading={isLoading}
      run={run}
      runHref={runHref}
    />
  )
}

export const RunbookConfigPanel = ({
  config,
  name,
  ...props
}: {
  config?: TRunbookConfig
  name: string
} & Omit<IPanel, 'heading' | 'children'>) => (
  <Panel
    heading={`${name} config`}
    panelKey={`runbook-config-${config?.runbook_id ?? name}`}
    size="half"
    {...props}
  >
    <RunbookConfigPanelBody config={config} />
  </Panel>
)
