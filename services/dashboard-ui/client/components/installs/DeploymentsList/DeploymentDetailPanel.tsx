import { Badge } from '@/components/common/Badge'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TInstallDeploymentRecord } from '@/types'
import { humanize } from '@/utils/string-utils'

const CHANGE_THEME = {
  add: 'success',
  remove: 'error',
  change: 'warn',
} as const

const CHANGE_PREFIX = {
  add: '+',
  remove: '-',
  change: '~',
} as const

const ChangeRows = ({
  changes,
}: {
  changes: TInstallDeploymentRecord['change_groups'][number]['changes']
}) => (
  <div className="flex flex-col divide-y border rounded-md overflow-hidden">
    {changes.map((change, index) => (
      <div
        key={`${change.path}-${change.operation}-${index}`}
        className="grid grid-cols-[1rem_minmax(0,1fr)] md:grid-cols-[1rem_minmax(10rem,1fr)_minmax(0,2fr)] items-center gap-3 px-3 py-2"
      >
        <Text
          variant="subtext"
          family="mono"
          weight="strong"
          theme={CHANGE_THEME[change.operation]}
        >
          {CHANGE_PREFIX[change.operation]}
        </Text>
        <Text variant="subtext" family="mono" weight="strong">
          {change.path}
        </Text>
        <div className="col-start-2 md:col-start-auto flex items-center gap-2 min-w-0">
          {change.is_redacted ? (
            <Badge size="sm" theme="neutral">
              Redacted
            </Badge>
          ) : (
            <>
              {change.previous_value !== undefined && (
                <Text
                  variant="subtext"
                  family="mono"
                  theme="neutral"
                  className="truncate"
                >
                  {change.previous_value}
                </Text>
              )}
              {change.previous_value !== undefined &&
                change.next_value !== undefined && (
                  <Icon
                    variant="ArrowRightIcon"
                    size={12}
                    className="shrink-0 text-cool-grey-400"
                  />
                )}
              {change.next_value !== undefined && (
                <Text variant="subtext" family="mono" className="truncate">
                  {change.next_value}
                </Text>
              )}
            </>
          )}
        </div>
      </div>
    ))}
  </div>
)

const AffectedResourceBadges = ({
  affectedResources,
}: {
  affectedResources: TInstallDeploymentRecord['affected_resources']
}) => (
  <div className="flex flex-wrap gap-2">
    {affectedResources.stack && (
      <Badge size="sm" theme="neutral">
        Stack
      </Badge>
    )}
    {affectedResources.sandbox && (
      <Badge size="sm" theme="neutral">
        Sandbox
      </Badge>
    )}
    {affectedResources.components.map((component) => (
      <Badge key={component} size="sm" variant="code" theme="neutral">
        {component}
      </Badge>
    ))}
    {affectedResources.images.map((image) => (
      <Badge key={image} size="sm" variant="code" theme="neutral">
        {image}
      </Badge>
    ))}
  </div>
)

export interface IDeploymentDetailPanel extends IPanel {
  deployment: TInstallDeploymentRecord
  orgId: string
  appId: string
  installId: string
}

export const DeploymentDetailPanel = ({
  deployment,
  orgId,
  appId,
  installId,
  ...props
}: IDeploymentDetailPanel) => {
  const branchHref = deployment.app_branch
    ? `/${orgId}/apps/${appId}/branches/${deployment.app_branch.id}`
    : undefined
  const workflowHref = deployment.workflow
    ? `/${orgId}/installs/${installId}/history/${deployment.workflow.id}`
    : undefined

  return (
    <Panel heading={deployment.title} size="half" {...props}>
      <div className="flex items-center gap-2 flex-wrap">
        <Status status={deployment.status} variant="badge" />
        <Badge size="sm" theme="neutral">
          {humanize(deployment.type)}
        </Badge>
      </div>

      {deployment.summary && (
        <Text variant="subtext" theme="neutral">
          {deployment.summary}
        </Text>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <LabeledValue label="Applied">
          <Time
            time={deployment.created_at}
            format="relative"
            variant="subtext"
          />
        </LabeledValue>
        <LabeledValue label="Deployment ID">
          <ID>{deployment.id}</ID>
        </LabeledValue>
        {deployment.app_branch && branchHref && (
          <LabeledValue label="App branch">
            <span className="flex items-center gap-2 flex-wrap">
              <Link href={branchHref}>{deployment.app_branch.name}</Link>
              {deployment.app_branch.sha && (
                <Badge size="sm" variant="code" theme="neutral">
                  {deployment.app_branch.sha.slice(0, 8)}
                </Badge>
              )}
            </span>
          </LabeledValue>
        )}
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
      <AffectedResourceBadges
        affectedResources={deployment.affected_resources}
      />

      {deployment.change_groups.map((group) => (
        <div key={group.id} className="flex flex-col gap-3">
          <Divider dividerWord={group.label} />
          {group.summary && (
            <Text variant="subtext" theme="neutral">
              {group.summary}
            </Text>
          )}
          {group.changes.length > 0 && <ChangeRows changes={group.changes} />}
          {group.file_diff && (
            <CodeBlock language={group.diff_language ?? 'diff'} showCopy>
              {group.file_diff}
            </CodeBlock>
          )}
        </div>
      ))}
    </Panel>
  )
}
