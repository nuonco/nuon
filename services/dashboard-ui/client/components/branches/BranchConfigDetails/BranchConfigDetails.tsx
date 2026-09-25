import { AppInputs } from '@/components/apps/config/AppInputs'
import { AppKubernetesContexts } from '@/components/apps/config/AppKubernetesContexts'
import { AppRunner } from '@/components/apps/config/AppRunner'
import { AppSandbox } from '@/components/apps/config/AppSandbox'
import { AppStack } from '@/components/apps/config/AppStack'
import { Badge } from '@/components/common/Badge'
import { Divider } from '@/components/common/Divider'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Loading } from '@/components/common/Loading'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TAppConfig } from '@/types'

export interface IBranchConfigDetails extends IPanel {
  config: TAppConfig
  fullConfig?: TAppConfig
  isLoading?: boolean
}

export const BranchConfigDetails = ({
  config,
  fullConfig,
  isLoading,
  ...props
}: IBranchConfigDetails) => {
  const contents = fullConfig ?? config
  const commit = config?.vcs_connection_commit

  return (
    <Panel
      heading={config?.version ? `Config v${config.version}` : 'App config'}
      size="half"
      {...props}
    >
      <div className="flex items-center gap-2">
        <Status
          variant="badge"
          status={config?.status_v2?.status || config?.status || 'unknown'}
        />
        <ID>{config?.id}</ID>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <LabeledValue label="Created">
          {config?.created_at ? (
            <Time
              variant="subtext"
              time={config.created_at}
              format="relative"
            />
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
        <LabeledValue label="CLI version">
          {config?.cli_version || <Icon variant="MinusIcon" />}
        </LabeledValue>
        <LabeledValue label="Components">
          <Text variant="subtext" family="mono">
            {config?.component_ids?.length ?? 0}
          </Text>
        </LabeledValue>
        <LabeledValue label="Actions">
          <Text variant="subtext" family="mono">
            {config?.action_ids?.length ?? 0}
          </Text>
        </LabeledValue>
        <LabeledValue label="Runbooks">
          <Text variant="subtext" family="mono">
            {config?.runbook_ids?.length ?? 0}
          </Text>
        </LabeledValue>
        <LabeledValue label="Checksum">
          {config?.checksum ? (
            <Text variant="subtext" family="mono">
              {config.checksum.slice(0, 12)}
            </Text>
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
      </div>

      {commit ? (
        <LabeledValue label="Commit">
          <span className="flex flex-col gap-1">
            {commit.message ? (
              <Text variant="body" weight="strong">
                {commit.message.split('\n')[0]?.trim()}
              </Text>
            ) : null}
            <span className="flex flex-wrap items-center gap-2">
              {commit.sha ? (
                <Badge size="sm" variant="code">
                  {commit.sha.slice(0, 8)}
                </Badge>
              ) : null}
              {commit.author_name ? (
                <Text variant="subtext" theme="neutral">
                  {commit.author_name}
                </Text>
              ) : null}
            </span>
          </span>
        </LabeledValue>
      ) : null}

      {isLoading && !fullConfig ? (
        <Loading variant="large" />
      ) : (
        <>
          {contents?.sandbox ? (
            <>
              <Divider dividerWord="Sandbox" />
              <AppSandbox appConfig={contents} />
            </>
          ) : null}

          {contents?.stack ? (
            <>
              <Divider dividerWord="Stack" />
              <AppStack appConfig={contents} />
            </>
          ) : null}

          {contents?.runner ? (
            <>
              <Divider dividerWord="Runner" />
              <AppRunner appConfig={contents} />
            </>
          ) : null}

          {contents?.input ? (
            <>
              <Divider dividerWord="Inputs" />
              <AppInputs appConfig={contents} />
            </>
          ) : null}

          {contents?.kubernetes_contexts?.contexts?.length ? (
            <>
              <Divider dividerWord="Kubernetes contexts" />
              <AppKubernetesContexts appConfig={contents} />
            </>
          ) : null}
        </>
      )}
    </Panel>
  )
}
