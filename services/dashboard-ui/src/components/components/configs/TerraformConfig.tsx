'use client'

import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TKeyValue } from '@/types'

interface TerraformVariablesFilesModalProps extends IModal {
  variablesFiles: string[]
}

interface TerraformVariablesModalProps extends IModal {
  variables: Record<string, string>
}

export const TerraformVariablesFilesModal = ({
  variablesFiles,
  ...props
}: TerraformVariablesFilesModalProps) => {
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
            This is the Terraform variables files for this component configuration.
          </Text>
          <div className="flex justify-end">
            <ClickToCopyButton textToCopy={variablesFilesContent} className="w-fit" />
          </div>
        </div>
        <CodeBlock language="hcl" className="!max-h-fit">{variablesFilesContent}</CodeBlock>
      </div>
    </Modal>
  )
}

export const TerraformVariablesModal = ({
  variables,
  ...props
}: TerraformVariablesModalProps) => {
  const keyValuePairs: TKeyValue[] = Object.entries(variables).map(([key, value]) => ({
    key,
    value,
    type: 'string'
  }))

  const variablesText = Object.entries(variables)
    .map(([key, value]) => `${key} = "${value}"`)
    .join('\n')

  return (
    <Modal
      heading={
        <Text variant="h3" weight="strong" className="!flex items-center gap-2">
          <Icon variant="List" size="20" />
          Terraform Variables
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
            This is the Terraform variables for this component configuration.
          </Text>
          <div className="flex justify-end">
            <ClickToCopyButton textToCopy={variablesText} className="w-fit" />
          </div>
        </div>
        <KeyValueList values={keyValuePairs} />
      </div>
    </Modal>
  )
}
