import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { GitRepo } from '@/components/common/GitRepo'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
import { objectToKeyValueArray } from '@/utils/data-utils'
import type { TActionConfigStep } from '@/types'

export interface IActionStep {
  index: number
  step: TActionConfigStep
}

export const ActionStep = ({ index, step }: IActionStep) => {
  return (
    <Expand
      className="border rounded-md"
      headerClassName=""
      heading={
        <Text weight="strong">
          {index + 1}. {step.name}
        </Text>
      }
      id={`step-${index}`}
      isOpen
    >
      <div className="flex flex-col gap-8 p-4 border-t">
        {step.inline_contents?.length ? (
          <div className="flex flex-col gap-2">
            <Text weight="strong">
              Inline contents
            </Text>
            <CodeBlock language="bash">{step.inline_contents}</CodeBlock>
          </div>
        ) : null}

        {step.command?.length ? (
          <div className="flex flex-col gap-2">
            <Text weight="strong">
              Command
            </Text>
            <CodeBlock language="bash">{step.command}</CodeBlock>
          </div>
        ) : null}

        {step.connected_github_vcs_config ? (
          <div className="flex flex-col gap-2">
            <Text weight="strong">
              GitHub VCS
            </Text>
            <GitRepo vcsConfig={step.connected_github_vcs_config} />
          </div>
        ) : null}

        {step.public_git_vcs_config ? (
          <div className="flex flex-col gap-2">
            <Text weight="strong">
              Git VCS
            </Text>
            <GitRepo vcsConfig={step.public_git_vcs_config} />
          </div>
        ) : null}

        {step.env_vars && Object.keys(step.env_vars).length ? (
          <div className="flex flex-col gap-2">
            <Text weight="strong">
              Environment variables
            </Text>
            <KeyValueList values={objectToKeyValueArray(step.env_vars)} />
          </div>
        ) : null}
      </div>
    </Expand>
  )
}
