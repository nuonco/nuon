import { ModalStory } from '@/components/__stories__/helpers'
import { DeleteOrgLinkModal } from './DeleteOrgLink'

export default { title: 'Slack/DeleteOrgLink' }

export const Default = () => (
  <ModalStory>
    <DeleteOrgLinkModal
      teamId="T0123456789"
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <DeleteOrgLinkModal
      teamId="T0123456789"
      isPending={true}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)
