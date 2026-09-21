import { Badge } from '@/components/common/Badge'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Divider } from '@/components/common/Divider'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { humanize } from '@/utils/string-utils'
import { ConfigurationChangeRows } from './ConfigurationChangeRows'
import { DeploymentAffectedResourceBadges } from './DeploymentAffectedResourceBadges'
import type {
  TDeploymentRecord,
  TDeploymentRecordType,
  TPlaygroundInstall,
} from './types'

export const DEPLOYMENT_TYPE_LABELS: Record<TDeploymentRecordType, string> = {
  provision: 'Provision',
  reprovision: 'Reprovision',
  sandbox_reprovision: 'Sandbox reprovision',
  app_branch_update: 'App branch update',
  component_deploy: 'Component deploy',
  image_update: 'Image update',
  stack_update: 'Stack update',
  install_config_update: 'Install config update',
}

export const DeploymentChangeDetails = ({
  deployment,
  install,
  ...props
}: {
  deployment: TDeploymentRecord
  install: TPlaygroundInstall
} & IPanel) => {
  const branchHref = `/${install.orgId}/apps/${install.appId}/branches/${deployment.appBranch.id}`
  const workflowHref = deployment.workflow
    ? `/${install.orgId}/installs/${install.id}/history/${deployment.workflow.id}`
    : undefined

  return (
    <Panel heading={deployment.title} size="half" {...props}>
      <div className="flex items-center gap-2 flex-wrap">
        <Status status={deployment.status} variant="badge" />
        <Badge size="sm" theme="neutral">
          {DEPLOYMENT_TYPE_LABELS[deployment.type]}
        </Badge>
      </div>

      <Text variant="subtext" theme="neutral">
        {deployment.summary}
      </Text>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <LabeledValue label="Applied">
          <Time
            time={deployment.createdAt}
            format="relative"
            variant="subtext"
          />
        </LabeledValue>
        <LabeledValue label="Deployment ID">
          <ID>{deployment.id}</ID>
        </LabeledValue>
        <LabeledValue label="App branch">
          <span className="flex items-center gap-2 flex-wrap">
            <Link href={branchHref}>{deployment.appBranch.name}</Link>
            {deployment.appBranch.sha && (
              <Badge size="sm" variant="code" theme="neutral">
                {deployment.appBranch.sha.slice(0, 8)}
              </Badge>
            )}
          </span>
        </LabeledValue>
        {deployment.workflow && workflowHref && (
          <LabeledValue label="Workflow">
            <span className="flex items-center gap-2 flex-wrap">
              <Link href={workflowHref}>{deployment.workflow.name}</Link>
              <Badge size="sm" theme="neutral">
                {humanize(deployment.workflow.type)}
              </Badge>
            </span>
          </LabeledValue>
        )}
      </div>

      <Divider dividerWord="Affected resources" />
      <DeploymentAffectedResourceBadges
        affectedResources={deployment.affectedResources}
      />

      {deployment.changeGroups.map((group) => (
        <div key={group.id} className="flex flex-col gap-3">
          <Divider dividerWord={group.label} />
          <Text variant="subtext" theme="neutral">
            {group.summary}
          </Text>
          {group.changes.length > 0 && (
            <ConfigurationChangeRows changes={group.changes} />
          )}
          {group.fileDiff && (
            <CodeBlock language={group.diffLanguage ?? 'diff'} showCopy>
              {group.fileDiff}
            </CodeBlock>
          )}
        </div>
      ))}
    </Panel>
  )
}
