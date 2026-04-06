import { CodeBlock } from '@/components/common/CodeBlock'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface ITerraformBackendConfigModal extends IModal {
  content: string
  onDownload: () => void
}

export const TerraformBackendConfigModal = ({
  content,
  onDownload,
  ...props
}: ITerraformBackendConfigModal) => {
  return (
    <Modal
      heading="Use Terraform CLI"
      primaryActionTrigger={{
        children: 'Download nuon_backend.tf',
        onClick: onDownload,
        variant: 'primary',
      }}
      size="half"
      {...props}
    >
      <div className="flex flex-col gap-4">
        <Text>
          Download the backend config and add it to your Terraform project. Then set your API token and initialize:
        </Text>
        <CodeBlock language="bash">{`export TF_HTTP_AUTHORIZATION="Bearer $(nuon orgs api-token -j | tr -d '\"')"
terraform init -reconfigure`}</CodeBlock>
        <CodeBlock language="hcl">{content}</CodeBlock>
      </div>
    </Modal>
  )
}
