import { Badge } from '@/components/common/Badge'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Duration } from '@/components/common/Duration'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { KeyValueList } from '@/components/common/KeyValueList'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TRunbookStep } from '@/lib/ctl-api/apps/runbooks/get-runbooks'
import { objectToKeyValueArray } from '@/utils/data-utils'

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
            variant={step.type === 'deploy' ? 'RocketIcon' : 'TerminalIcon'}
            size={16}
          />
          <Text weight="strong">
            {index + 1}. {step.name}
          </Text>
          <Badge variant="code" size="sm" theme="neutral">
            {step.type}
          </Badge>
        </span>
      }
      id={`step-${index}`}
      isOpen
    >
      <div className="flex flex-col gap-4 p-4 border-t">
        <div className="flex flex-wrap gap-4">
          {step.component_name ? (
            <LabeledValue label="Component">
              <Text variant="subtext">{step.component_name}</Text>
            </LabeledValue>
          ) : null}
          {step.type === 'deploy' ? (
            <LabeledValue label="Deploy dependencies">
              <Badge
                variant="code"
                size="sm"
                theme={step.deploy_dependencies ? 'info' : 'neutral'}
              >
                {step.deploy_dependencies ? 'Yes' : 'No'}
              </Badge>
            </LabeledValue>
          ) : null}
          {step.action_workflow_id ? (
            <LabeledValue label="Action ID">
              <span className="flex items-center gap-2">
                <ID>{step.action_workflow_id}</ID>
              </span>
            </LabeledValue>
          ) : null}
          {step.role ? (
            <LabeledValue label="Role">
              <Text variant="subtext">{step.role}</Text>
            </LabeledValue>
          ) : null}
          {step.timeout ? (
            <LabeledValue label="Timeout">
              <Duration nanoseconds={step.timeout} variant="subtext" />
            </LabeledValue>
          ) : null}
        </div>

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
          <Link
            className="text-sm"
            href={`${actionBasePath}/actions/${step.action_workflow_id}`}
          >
            View action
            <Icon variant="CaretRightIcon" />
          </Link>
        ) : null}
      </div>
    </Expand>
  )
}
