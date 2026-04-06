export default {
  title: 'VCS Connections/VCSConnectionSuccess',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { VCSConnectionSuccessModal } from './VCSConnectionSuccess'

const mockConnection = {
  id: 'vcs-abc123',
  github_account_name: 'acme-corp',
  github_account_id: '12345',
  github_install_id: 'inst-1',
} as any

export const Default = () => (
  <ModalStory>
    <VCSConnectionSuccessModal
      orgName="My Org"
      vcs_connection={mockConnection}
    />
  </ModalStory>
)
