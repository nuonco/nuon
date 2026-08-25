import { Badge } from '@/components/common/Badge'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Duration } from '@/components/common/Duration'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TRunbookStep } from '@/lib/ctl-api/apps/runbooks/get-runbooks'
import { objectToKeyValueArray } from '@/utils/data-utils'
import { humanize } from '@/utils/string-utils'

export interface IRunbookStep {
  index: number
  step: TRunbookStep
  actionBasePath?: string
}

export const RunbookStep = ({ index, step, actionBasePath }: IRunbookStep) => {
  return (
    <Expand
      className="border rounded-md"
      heading={
        <span className="flex items-center gap-2">
          <Icon
            variant={
              step.type === 'component_deploy' || step.type === 'deploy'
                ? 'RocketIcon'
                : step.type === 'component_tear_down'
                  ? 'TrashIcon'
                  : step.type === 'sandbox_reprovision' ||
                      step.type === 'sandbox_deprovision'
                    ? 'CubeIcon'
                    : 'TerminalIcon'
            }
            size={16}
          />
          <Text weight="strong">
            {index + 1}. {step.name}
          </Text>
          <Badge size="sm" theme="neutral">
            {humanize(step.type)}
          </Badge>
        </span>
      }
      id={`step-${index}`}
      isOpen
    >
      <div className="flex flex-col gap-4 p-4 border-t">
        <div className="flex flex-col divide-y">
          {step.component_name ? (
            <div className="flex items-center py-2 gap-4">
              <Text variant="subtext" theme="neutral" className="w-48 shrink-0">
                Component
              </Text>
              <Text variant="subtext">{step.component_name}</Text>
            </div>
          ) : null}
          {step.type === 'component_deploy' || step.type === 'deploy' ? (
            <div className="flex items-center py-2 gap-4">
              <Text variant="subtext" theme="neutral" className="w-48 shrink-0">
                Deploy dependents
              </Text>
              <Text variant="subtext">
                {step.deploy_dependents ? 'Yes' : 'No'}
              </Text>
            </div>
          ) : null}
          {step.type === 'component_tear_down' ? (
            <div className="flex items-center py-2 gap-4">
              <Text variant="subtext" theme="neutral" className="w-48 shrink-0">
                Tear down dependents
              </Text>
              <Text variant="subtext">
                {step.tear_down_dependents ? 'Yes' : 'No'}
              </Text>
            </div>
          ) : null}
          {step.type === 'sandbox_reprovision' ? (
            <div className="flex items-center py-2 gap-4">
              <Text variant="subtext" theme="neutral" className="w-48 shrink-0">
                Skip component deploys
              </Text>
              <Text variant="subtext">
                {step.skip_component_deploys ? 'Yes' : 'No'}
              </Text>
            </div>
          ) : null}
          {step.action_workflow_id ? (
            <div className="flex items-center py-2 gap-4">
              <Text variant="subtext" theme="neutral" className="w-48 shrink-0">
                Action ID
              </Text>
              <ID>{step.action_workflow_id}</ID>
            </div>
          ) : null}
          {step.role ? (
            <div className="flex items-center py-2 gap-4">
              <Text variant="subtext" theme="neutral" className="w-48 shrink-0">
                Role
              </Text>
              <Text variant="subtext">{step.role}</Text>
            </div>
          ) : null}
          {step.timeout ? (
            <div className="flex items-center py-2 gap-4">
              <Text variant="subtext" theme="neutral" className="w-48 shrink-0">
                Timeout
              </Text>
              <Duration nanoseconds={step.timeout} variant="subtext" />
            </div>
          ) : null}
          {step.type === 'wait_for_event' ? (
            <>
              <div className="flex items-center py-2 gap-4">
                <Text
                  variant="subtext"
                  theme="neutral"
                  className="w-48 shrink-0"
                >
                  Trigger
                </Text>
                <Text variant="subtext">
                  {step.trigger_name || step.trigger_id || 'Unknown'}
                </Text>
              </div>
              <div className="flex items-center py-2 gap-4">
                <Text
                  variant="subtext"
                  theme="neutral"
                  className="w-48 shrink-0"
                >
                  Event types
                </Text>
                <Text variant="subtext">
                  {step.event_types?.join(', ') || 'All event types'}
                </Text>
              </div>
              {!step.timeout ? (
                <div className="flex items-center py-2 gap-4">
                  <Text
                    variant="subtext"
                    theme="neutral"
                    className="w-48 shrink-0"
                  >
                    Timeout
                  </Text>
                  <Text variant="subtext">Wait indefinitely</Text>
                </div>
              ) : null}
            </>
          ) : null}
        </div>

        {step.type === 'wait_for_event' ? (
          <div className="flex flex-col gap-2">
            <Text weight="strong">Filters</Text>
            {step.filters?.length ? (
              step.filters.map((filter, filterIndex) => (
                <CodeBlock
                  key={`${filter.path}-${filterIndex}`}
                  language="text"
                >
                  {`${filter.from || 'payload'} ${filter.path || '—'} ${filter.op || '—'} ${JSON.stringify(filter.value)}`}
                </CodeBlock>
              ))
            ) : (
              <Text variant="subtext" theme="neutral">
                No filters. Every configured event type matches.
              </Text>
            )}
          </div>
        ) : null}

        {step.command ? (
          <div className="flex flex-col gap-2">
            <Text weight="strong">Command</Text>
            <CodeBlock language="bash">{step.command}</CodeBlock>
          </div>
        ) : null}

        {step.inline_contents ? (
          <div className="flex flex-col gap-2">
            <Text weight="strong">Inline contents</Text>
            <CodeBlock language="bash">{step.inline_contents}</CodeBlock>
          </div>
        ) : null}

        {step.env_vars && Object.keys(step.env_vars).length > 0 ? (
          <div className="flex flex-col gap-2">
            <Text weight="strong">Environment variables</Text>
            <KeyValueList values={objectToKeyValueArray(step.env_vars)} />
          </div>
        ) : null}
        {actionBasePath && step.action_workflow_id ? (
          <Link href={`${actionBasePath}/actions/${step.action_workflow_id}`}>
            View action
          </Link>
        ) : null}
      </div>
    </Expand>
  )
}
