import { Button } from '@/components/common/Button'
import { Card, type ICard } from '@/components/common/Card'
import { Cron } from '@/components/common/Cron'
import { GitRepo } from '@/components/common/GitRepo'
import { LabeledValue } from '@/components/common/LabeledValue'
import { OperationRolesList } from '@/components/common/OperationRolesList'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TSandboxConfig } from '@/types'

interface ISandboxConfigCard extends Omit<ICard, 'children'> {
  config: TSandboxConfig
  onViewEnvVars?: () => void
  onViewVariablesFiles?: () => void
}

export const SandboxConfigCard = ({
  config,
  onViewEnvVars,
  onViewVariablesFiles,
  ...props
}: ISandboxConfigCard) => {
  const hasEnvVars = config.env_vars && Object.keys(config.env_vars).length > 0
  const hasVariablesFiles =
    config.variables_files && config.variables_files.length > 0

  const vcsConfig =
    config.connected_github_vcs_config || config.public_git_vcs_config

  return (
    <Card {...props}>
      <div className="flex flex-col gap-6">
        <div className="flex items-center justify-between">
          <Text weight="strong">Configuration</Text>
          {(hasEnvVars || hasVariablesFiles) && (
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
            </div>
          )}
        </div>

        <div className="grid gap-6 grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          {config.terraform_version && (
            <LabeledValue label="Terraform version">
              {config.terraform_version}
            </LabeledValue>
          )}
          {config.aws_region_type && (
            <LabeledValue label="AWS region type">
              {config.aws_region_type}
            </LabeledValue>
          )}
          {config.drift_schedule && (
            <LabeledValue label="Drift schedule">
              <Cron cron={config.drift_schedule} variant="subtext" />
            </LabeledValue>
          )}
        </div>

        {config.operation_roles &&
          Object.keys(config.operation_roles).length > 0 && (
            <div className="flex flex-col gap-2">
              <Text variant="body" weight="strong" level={5}>
                Operation roles
              </Text>
              <OperationRolesList operationRoles={config.operation_roles} />
            </div>
          )}

        {vcsConfig && (
          <div className="pt-6 border-t">
            <div className="w-fit">
              <GitRepo vcsConfig={vcsConfig} isAutoGrid />
            </div>
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
