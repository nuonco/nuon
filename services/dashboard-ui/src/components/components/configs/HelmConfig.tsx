'use client'

import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TKeyValue } from '@/types'

interface HelmValuesFilesModalProps extends IModal {
  valuesFiles: string[]
}

interface HelmValuesModalProps extends IModal {
  values: Record<string, string>
}

export const HelmValuesFilesModal = ({
  valuesFiles,
  ...props
}: HelmValuesFilesModalProps) => {
  const valuesFilesContent = valuesFiles.join('\n---\n')

  return (
    <Modal
      heading={
        <Text variant="h3" weight="strong" className="!flex items-center gap-2">
          <Icon variant="FileCode" size="20" />
          Helm Values Files
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
            This is the Helm values files for this component configuration.
          </Text>
          <div className="flex justify-end">
            <ClickToCopyButton textToCopy={valuesFilesContent} className="w-fit" />
          </div>
        </div>
        <CodeBlock language="yaml">{valuesFilesContent}</CodeBlock>
      </div>
    </Modal>
  )
}

export const HelmValuesModal = ({
  values,
  ...props
}: HelmValuesModalProps) => {
  const keyValuePairs: TKeyValue[] = Object.entries(values).map(([key, value]) => ({
    key,
    value,
    type: 'string'
  }))

  const valuesText = Object.entries(values)
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n')

  return (
    <Modal
      heading={
        <Text variant="h3" weight="strong" className="!flex items-center gap-2">
          <Icon variant="List" size="20" />
          Helm Values
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
            This is the Helm values for this component configuration.
          </Text>
          <div className="flex justify-end">
            <ClickToCopyButton textToCopy={valuesText} className="w-fit" />
          </div>
        </div>
        <KeyValueList values={keyValuePairs} />
      </div>
    </Modal>
  )
}