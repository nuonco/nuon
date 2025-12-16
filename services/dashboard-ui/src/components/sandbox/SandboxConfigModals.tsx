'use client'

import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TKeyValue } from '@/types'

interface SandboxEnvironmentVariablesModalProps extends IModal {
  envVars: Record<string, string>
}

interface SandboxVariablesFilesModalProps extends IModal {
  variablesFiles: string[]
}

export const SandboxEnvironmentVariablesModal = ({
  envVars,
  ...props
}: SandboxEnvironmentVariablesModalProps) => {
  const envVarsList: TKeyValue[] = Object.entries(envVars).map(([key, value]) => ({
    key,
    value,
  }))

  const envVarsText = Object.entries(envVars)
    .map(([key, value]) => `${key}=${value}`)
    .join('\n')

  return (
    <Modal
      heading={
        <Text variant="h3" weight="strong" className="!flex items-center gap-2">
          <Icon variant="List" size="20" />
          Environment Variables
        </Text>
      }
      size="3/4"
      className="!max-h-[80vh]"
      childrenClassName="overflow-y-auto"
      {...props}
    >
      <div className="flex flex-col gap-4">
        <div className="flex justify-between items-center gap-4">
          <Text variant="body">
            Environment variables configured for this sandbox.
          </Text>
          <div className="flex justify-end">
            <ClickToCopyButton textToCopy={envVarsText} className="w-fit" />
          </div>
        </div>
        <KeyValueList values={envVarsList} />
      </div>
    </Modal>
  )
}

export const SandboxVariablesFilesModal = ({
  variablesFiles,
  ...props
}: SandboxVariablesFilesModalProps) => {
  const variablesFilesContent = variablesFiles.join('\n---\n')

  return (
    <Modal
      heading={
        <Text variant="h3" weight="strong" className="!flex items-center gap-2">
          <Icon variant="FileCode" size="20" />
          Terraform Variables Files
        </Text>
      }
      size="3/4"
      className="!max-h-[80vh]"
      childrenClassName="overflow-y-auto"
      {...props}
    >
      <div className="flex flex-col gap-4">
        <div className="flex justify-between items-center gap-4">
          <Text variant="body">
            Terraform variables files configured for this sandbox.
          </Text>
          <div className="flex justify-end">
            <ClickToCopyButton textToCopy={variablesFilesContent} className="w-fit" />
          </div>
        </div>
        <CodeBlock language="hcl" className="!max-h-fit">
          {variablesFilesContent}
        </CodeBlock>
      </div>
    </Modal>
  )
}