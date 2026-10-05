import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Cron } from '@/components/common/Cron'
import { EmptyState } from '@/components/common/EmptyState'
import { Hash } from '@/components/common/Hash'
import { Loading } from '@/components/common/Loading'
import { OperationRolesList } from '@/components/common/OperationRolesList'
import { useComponentConfigButtons } from '@/components/components/ComponentConfigCard'
import { ComponentType } from '@/components/components/ComponentType'
import { SignatureVerification } from '@/components/components/configs/SignatureVerification'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig } from '@/lib'
import type { TComponentConfig } from '@/types'
import { getComponentConfigDisplayData } from '@/utils/component-config-display'
import { ConfigProperties, ConfigSection } from './ConfigLayout'
import {
  InstallConfigSource,
  type IInstallConfigSource,
} from './InstallConfigSource'
import { useInstallConfigRun } from './use-install-config-run'

export interface IComponentConfigDetails extends IInstallConfigSource {
  config?: TComponentConfig
  emptyMessage: string
  isLoadingConfig?: boolean
}

const ComponentConfigBody = ({ config }: { config: TComponentConfig }) => {
  const buttons = useComponentConfigButtons(config)
  const { commonFields, typeSpecificFields, vcsInfo, operationRoles } =
    getComponentConfigDisplayData(config)

  return (
    <>
      <ConfigSection
        title="Configuration"
        actions={buttons.map((button) => (
          <Button
            key={button.label}
            onClick={button.onClick}
            size="sm"
            variant="secondary"
          >
            {button.label}
          </Button>
        ))}
      >
        <ConfigProperties
          properties={[
            ...typeSpecificFields.map(({ label, value }) => ({ label, value })),
            { label: 'Build timeout', value: commonFields.buildTimeout },
            { label: 'Deploy timeout', value: commonFields.deployTimeout },
            {
              label: 'Drift schedule',
              value: commonFields.driftSchedule ? (
                <Cron cron={commonFields.driftSchedule} variant="subtext" />
              ) : undefined,
            },
            {
              label: 'Checksum',
              value: commonFields.checksum ? (
                <Hash hash={commonFields.checksum} />
              ) : undefined,
            },
          ]}
        />
      </ConfigSection>
      {vcsInfo?.repo ? (
        <ConfigSection title="Repository">
          <ConfigProperties
            properties={[
              { label: 'Repository', value: vcsInfo.repo },
              { label: 'Branch', value: vcsInfo.branch },
              { label: 'Directory', value: vcsInfo.directory },
            ]}
          />
        </ConfigSection>
      ) : null}
      {operationRoles && Object.keys(operationRoles).length ? (
        <ConfigSection title="Operation roles">
          <OperationRolesList operationRoles={operationRoles} />
        </ConfigSection>
      ) : null}
      <SignatureVerification
        verification={config.external_image?.verification}
      />
    </>
  )
}

export const ComponentConfigDetails = ({
  config,
  emptyMessage,
  isLoadingConfig,
  behind,
  ...source
}: IComponentConfigDetails) => (
  <>
    <div className="flex flex-wrap items-center gap-3">
      {config?.type ? (
        <ComponentType type={config.type} variant="subtext" />
      ) : null}
      {config?.version ? (
        <Badge size="sm" theme="neutral" variant="code">
          v{config.version}
        </Badge>
      ) : null}
    </div>
    <InstallConfigSource behind={behind} {...source} />
    {isLoadingConfig ? (
      <Loading />
    ) : config ? (
      <ComponentConfigBody config={config} />
    ) : (
      <EmptyState
        variant="table"
        emptyTitle="No configuration"
        emptyMessage={emptyMessage}
      />
    )}
  </>
)

const useResourceAppConfig = (configId?: string) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { appConfig: pinned, isLoading: pinnedLoading } = useInstallAppConfig()
  const pinnedId = install?.app_config_id
  const usePinned = !configId || configId === pinnedId

  const other = useQuery({
    queryKey: ['app-config', org?.id, install?.app_id, configId, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: org!.id,
        appId: install!.app_id!,
        appConfigId: configId!,
        recurse: true,
      }),
    enabled: !usePinned && !!org?.id && !!install?.app_id && !!configId,
  })

  return {
    appConfig: usePinned ? pinned : other.data,
    isLoading: usePinned ? pinnedLoading : other.isLoading,
  }
}

const ComponentConfigPanelBody = ({
  appliedConfigId,
  componentId,
  emptyMessage,
}: {
  appliedConfigId?: string
  componentId: string
  emptyMessage: string
}) => {
  const { install } = useInstall()
  const configId = appliedConfigId || install?.app_config_id
  const behind =
    !!appliedConfigId &&
    !!install?.app_config_id &&
    appliedConfigId !== install.app_config_id
  const { isLoading, run, runHref } = useInstallConfigRun(configId)
  const { appConfig, isLoading: isLoadingConfig } =
    useResourceAppConfig(configId)
  const config = appConfig?.component_config_connections?.find(
    (connection) => connection.component_id === componentId
  )

  return (
    <ComponentConfigDetails
      behind={behind}
      config={config}
      emptyMessage={emptyMessage}
      isLoading={isLoading}
      isLoadingConfig={isLoadingConfig}
      run={run}
      runHref={runHref}
    />
  )
}

export const ComponentConfigPanel = ({
  appliedConfigId,
  componentId,
  emptyMessage = 'This component is not in the app config this install is using.',
  name,
  ...props
}: {
  appliedConfigId?: string
  componentId: string
  emptyMessage?: string
  name: string
} & Omit<IPanel, 'heading' | 'children'>) => (
  <Panel
    heading={`${name} config`}
    panelKey={`component-config-${componentId}`}
    size="half"
    {...props}
  >
    <ComponentConfigPanelBody
      appliedConfigId={appliedConfigId}
      componentId={componentId}
      emptyMessage={emptyMessage}
    />
  </Panel>
)
