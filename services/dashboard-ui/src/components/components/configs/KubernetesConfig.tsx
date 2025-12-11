'use client'

import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface KubernetesManifestModalProps extends IModal {
  manifest: string
}

export const KubernetesManifestModal = ({
  manifest,
  ...props
}: KubernetesManifestModalProps) => (
  <Modal
    heading={
      <Text variant="h3" weight="strong" className="!flex items-center gap-2">
        <Icon variant="FileCode" size="20" />
        Kubernetes Manifest
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
          This is the Kubernetes manifest file for this component configuration.
        </Text>
        <div className="flex justify-end">
          <ClickToCopyButton textToCopy={manifest} className="w-fit" />
        </div>
      </div>
      <CodeBlock language="yaml">{manifest}</CodeBlock>
    </div>
  </Modal>
)
