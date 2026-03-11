import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useConfig } from '@/hooks/use-config'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { downloadFileOnClick } from '@/utils/file-download'

const getBackendContent = (apiUrl: string, workspaceId: string, orgId: string) => `terraform {
  backend "http" {
    lock_method    = "POST"
    unlock_method  = "POST"
    address        = "${apiUrl}/v1/terraform-backend?workspace_id=${workspaceId}&org_id=${orgId}"
    lock_address   = "${apiUrl}/v1/terraform-workspaces/${workspaceId}/lock?org_id=${orgId}"
    unlock_address = "${apiUrl}/v1/terraform-workspaces/${workspaceId}/unlock?org_id=${orgId}"
  }
}
`

interface ITerraformBackendConfig {
  workspaceId: string
}

export const TerraformBackendConfigModal = ({
  workspaceId,
  ...props
}: ITerraformBackendConfig & IModal) => {
  const config = useConfig()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()

  const content = getBackendContent(config.apiUrl, workspaceId, org.id)

  return (
    <Modal
      heading="Use Terraform CLI"
      primaryActionTrigger={{
        children: 'Download nuon_backend.tf',
        onClick: () =>
          downloadFileOnClick({
            content,
            filename: 'nuon_backend.tf',
            callback: () => removeModal(props.modalId),
          }),
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

export const TerraformBackendConfigButton = ({
  workspaceId,
  ...props
}: ITerraformBackendConfig & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <TerraformBackendConfigModal workspaceId={workspaceId} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      {props.children}
    </Button>
  )
}
