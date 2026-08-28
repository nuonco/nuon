import { Badge } from '@/components/common/Badge'
import { CodeBlock } from '@/components/common/CodeBlock'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { diffLines } from '@/utils/code-utils'

type StackPropertyChange = {
  path?: string
  before?: unknown
  after?: unknown
}

type StackChange = {
  action?: string
  logical_resource_id?: string
  resource_type?: string
  replacement?: string
  scope?: string[]
  property_changes?: StackPropertyChange[]
}

type StackPlan = {
  stack_name?: string
  status?: string
  status_reason?: string
  no_op?: boolean
  changes?: StackChange[]
}

const actionTheme = (action?: string) => {
  if (action === 'Add') return 'success'
  if (action === 'Remove') return 'error'
  return 'warn'
}

const serialized = (value: unknown) =>
  value === undefined ? undefined : JSON.stringify(value, null, 2)

export const SnapshotCloudFormationPlan = ({ plan }: { plan: unknown }) => {
  const stackPlan = plan as StackPlan
  const changes = stackPlan.changes ?? []

  if (stackPlan.no_op || changes.length === 0) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No stack changes"
        emptyMessage="The CloudFormation change set did not contain any resource changes."
      />
    )
  }

  return (
    <div className="flex flex-col gap-3">
      <HeadingGroup>
        <span className="flex flex-wrap items-center gap-2">
          <Text weight="strong">CloudFormation change set</Text>
          <Badge theme="neutral">{changes.length} resources</Badge>
        </span>
        {stackPlan.stack_name ? (
          <Text variant="subtext" theme="neutral" family="mono">
            {stackPlan.stack_name}
          </Text>
        ) : null}
        {stackPlan.status_reason ? (
          <Text variant="subtext" theme="error">
            {stackPlan.status_reason}
          </Text>
        ) : null}
      </HeadingGroup>
      {changes.map((change, index) => (
        <Expand
          id={`captured-stack-change-${change.logical_resource_id ?? index}`}
          isOpen
          className="border rounded-md"
          headerClassName="px-4 py-3"
          heading={
            <div className="flex flex-1 items-center justify-between gap-4 text-left">
              <span className="flex flex-col min-w-0">
                <Text weight="strong">
                  {change.logical_resource_id ?? 'Stack resource'}
                </Text>
                <Text variant="subtext" theme="neutral" family="mono">
                  {change.resource_type ?? 'CloudFormation resource'}
                </Text>
              </span>
              <span className="flex items-center gap-2">
                {change.replacement && change.replacement !== 'False' ? (
                  <Badge theme="warn">Replacement {change.replacement}</Badge>
                ) : null}
                <Badge theme={actionTheme(change.action)}>
                  {change.action ?? 'Modify'}
                </Badge>
              </span>
            </div>
          }
        >
          <div className="flex flex-col gap-3 border-t p-4">
            {change.scope?.length ? (
              <Text variant="subtext" theme="neutral">
                Scope: {change.scope.join(', ')}
              </Text>
            ) : null}
            {change.property_changes?.length ? (
              change.property_changes.map((property, propertyIndex) => (
                <div
                  className="flex flex-col gap-2"
                  key={`${property.path}-${propertyIndex}`}
                >
                  <Text variant="label" family="mono">
                    {property.path ?? 'Property'}
                  </Text>
                  <CodeBlock language="json" isDiff>
                    {diffLines(
                      serialized(property.before),
                      serialized(property.after)
                    )}
                  </CodeBlock>
                </div>
              ))
            ) : (
              <Text variant="subtext" theme="neutral">
                CloudFormation reported a resource-level change without property
                values.
              </Text>
            )}
          </div>
        </Expand>
      ))}
    </div>
  )
}
