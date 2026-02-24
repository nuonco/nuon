import { GitRepo } from '@/components/common/GitRepo'
import { KeyValueList } from '@/components/common/KeyValueList'
import { LabeledValue } from '@/components/common/LabeledValue'
import { OperationRolesList } from '@/components/common/OperationRolesList'
import { Text } from '@/components/common/Text'
import type { TAppConfig } from '@/types'
import { objectToKeyValueArray } from '@/utils/data-utils'

export interface IAppSandbox {
  appConfig: TAppConfig
}

export const AppSandbox = ({ appConfig }: IAppSandbox) => {
  const sandboxConfig = appConfig?.sandbox
  const sandboxConfigRepo =
    sandboxConfig?.connected_github_vcs_config ||
    sandboxConfig?.public_git_vcs_config ||
    {}
  const sandboxVariables = objectToKeyValueArray(sandboxConfig?.variables)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex gap-6 items-start justify-start">
        <GitRepo vcsConfig={sandboxConfigRepo} />
        <LabeledValue label="Terraform">
          {sandboxConfig?.terraform_version}
        </LabeledValue>
      </div>
      {sandboxVariables?.length ? (
        <div>
          <Text variant="subtext" weight="strong">
            Variables
          </Text>

          <KeyValueList
            emptyStateProps={{
              emptyTitle: 'No sandbox variables',
              emptyMessage: 'No variables configured for this sandbox',
            }}
            values={sandboxVariables}
          />
        </div>
      ) : null}
      {sandboxConfig?.operation_roles &&
        Object.keys(sandboxConfig.operation_roles).length > 0 && (
          <div>
            <Text variant="subtext" weight="strong">
              Operation Roles
            </Text>
            <OperationRolesList
              operationRoles={sandboxConfig.operation_roles}
            />
          </div>
        )}
    </div>
  )
}
