import { Button } from '@/components/common/Button'
import { Card, type ICard } from '@/components/common/Card'
import { Cron } from '@/components/common/Cron'
import { GitRepo } from '@/components/common/GitRepo'
import { KeyValueList } from '@/components/common/KeyValueList'
import { LabeledValue } from '@/components/common/LabeledValue'
import { OperationRolesList } from '@/components/common/OperationRolesList'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TSandboxConfig } from '@/types'
import { objectToKeyValueArray } from '@/utils/data-utils'

interface ISandboxConfigCard extends Omit<ICard, 'children'> {
  config: TSandboxConfig
  onViewEnvVars?: () => void
  onViewVariablesFiles?: () => void
  onViewPulumiConfig?: () => void
}

export const SandboxConfigCard = ({
  config,
  onViewEnvVars,
  onViewVariablesFiles,
  onViewPulumiConfig,
  ...props
}: ISandboxConfigCard) => {
  const isPulumi = config.type === 'pulumi'
  const hasEnvVars = config.env_vars && Object.keys(config.env_vars).length > 0
  const hasVariablesFiles =
    config.variables_files && config.variables_files.length > 0
  const hasPulumiConfig =
    isPulumi && config.pulumi_config && Object.keys(config.pulumi_config).length > 0
  const sandboxVariables = objectToKeyValueArray(config.variables)

  const vcsConfig =
    config.connected_github_vcs_config || config.public_git_vcs_config

  return (
    <Card {...props}>
      <div className="flex flex-col gap-6">
        <div className="flex items-center justify-between">
          <Text weight="strong">Configuration</Text>
          <div className="flex gap-2">
            {hasEnvVars && onViewEnvVars && (
              <Button
                variant="secondary"
                size="sm"
                onClick={onViewEnvVars}
              >
                View env vars
              </Button>
            )}
            {hasVariablesFiles && onViewVariablesFiles && (
              <Button
                variant="secondary"
                size="sm"
                onClick={onViewVariablesFiles}
              >
                View variables files
              </Button>
            )}
            {hasPulumiConfig && onViewPulumiConfig && (
              <Button
                variant="secondary"
                size="sm"
                onClick={onViewPulumiConfig}
              >
                View Pulumi config
              </Button>
            )}
          </div>
        </div>

        <div className="flex gap-6 items-start justify-start flex-wrap">
          {vcsConfig && <GitRepo vcsConfig={vcsConfig} />}
          {isPulumi ? (
            <>
              <LabeledValue label="Type">Pulumi</LabeledValue>
              {config.runtime && (
                <LabeledValue label="Runtime">{config.runtime}</LabeledValue>
              )}
              {config.pulumi_version && (
                <LabeledValue label="Pulumi version">
                  {config.pulumi_version}
                </LabeledValue>
              )}
            </>
          ) : (
            config.terraform_version && (
              <LabeledValue label="Terraform">
                {config.terraform_version}
              </LabeledValue>
            )
          )}
          {config.drift_schedule && (
            <LabeledValue label="Drift schedule">
              <Cron cron={config.drift_schedule} variant="subtext" />
            </LabeledValue>
          )}
        </div>

        {sandboxVariables?.length ? (
          <div>
            <Text variant="subtext" weight="strong">
              Variables
            </Text>
            <KeyValueList values={sandboxVariables} />
          </div>
        ) : null}

        {config.operation_roles &&
          Object.keys(config.operation_roles).length > 0 && (
            <div>
              <Text variant="subtext" weight="strong">
                Operation roles
              </Text>
              <OperationRolesList operationRoles={config.operation_roles} />
            </div>
          )}
      </div>
    </Card>
  )
}

export const SandboxConfigCardSkeleton = (props: Omit<ICard, 'children'>) => {
  return (
    <Card {...props}>
      <div className="flex flex-col gap-6">
        <div className="flex items-center gap-4">
          <Skeleton height="24px" width="150px" />
        </div>

        <div className="flex gap-6 items-start flex-wrap">
          <LabeledValue label={<Skeleton height="17px" width="120px" />}>
            <Skeleton height="23px" width="80px" />
          </LabeledValue>

          <LabeledValue label={<Skeleton height="17px" width="100px" />}>
            <Skeleton height="23px" width="60px" />
          </LabeledValue>
        </div>
      </div>
    </Card>
  )
}
