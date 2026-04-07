export default {
  title: 'Terraform/TerraformBackendConfig',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { TerraformBackendConfigModal } from './TerraformBackendConfig'

const mockContent = `terraform {
  backend "http" {
    lock_method    = "POST"
    unlock_method  = "POST"
    address        = "https://api.nuon.co/v1/terraform-backend?workspace_id=ws-123&org_id=org-456"
    lock_address   = "https://api.nuon.co/v1/terraform-workspaces/ws-123/lock?org_id=org-456"
    unlock_address = "https://api.nuon.co/v1/terraform-workspaces/ws-123/unlock?org_id=org-456"
  }
}
`

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <TerraformBackendConfigModal
      content={mockContent}
      onDownload={noop}
    />
  </ModalStory>
)
