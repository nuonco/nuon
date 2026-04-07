export default {
  title: 'Team/ResendOrgInvite',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ResendOrgInviteModal } from './ResendOrgInvite'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <ResendOrgInviteModal
      email="pending@example.com"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <ResendOrgInviteModal
      email="pending@example.com"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ResendOrgInviteModal
      email="pending@example.com"
      isPending={false}
      error={{ error: 'Rate limit exceeded', description: '', user_error: true }}
      onSubmit={noop}
    />
  </ModalStory>
)
