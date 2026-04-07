export default {
  title: 'Terraform/UnlockTerraformWorkspace',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { UnlockTerraformWorkspaceModal } from './UnlockTerraformWorkspace'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <UnlockTerraformWorkspaceModal
      description="sandbox workspace"
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <UnlockTerraformWorkspaceModal
      description="sandbox workspace"
      isPending={true}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <UnlockTerraformWorkspaceModal
      description="sandbox workspace"
      isPending={false}
      error={{ error: 'Workspace is not locked' } as any}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)
