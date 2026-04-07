import { Button, type IButtonAsButton } from '@/components/common/Button'
import type { IModal } from '@/components/surfaces/Modal'
import { useConfig } from '@/hooks/use-config'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { downloadFileOnClick } from '@/utils/file-download'
import { TerraformBackendConfigModal } from './TerraformBackendConfig'

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

export const TerraformBackendConfigModalContainer = ({
  workspaceId,
  ...props
}: ITerraformBackendConfig & IModal) => {
  const config = useConfig()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()

  const content = getBackendContent(config.apiUrl, workspaceId, org.id)

  return (
    <TerraformBackendConfigModal
      content={content}
      onDownload={() =>
        downloadFileOnClick({
          content,
          filename: 'nuon_backend.tf',
          callback: () => removeModal(props.modalId),
        })
      }
      {...props}
    />
  )
}

export const TerraformBackendConfigButton = ({
  workspaceId,
  ...props
}: ITerraformBackendConfig & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <TerraformBackendConfigModalContainer workspaceId={workspaceId} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      {props.children}
    </Button>
  )
}
