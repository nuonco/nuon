export default {
  title: 'Stacks/SendStackOutputsModal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { SendStackOutputsModal } from './SendStackOutputsModal'

export const Default = () => (
  <ModalStory>
    <SendStackOutputsModal
      phoneHomeId="ph-123"
      versionId="ver-abc"
      onSend={() => {}}
      isPending={false}
      error={undefined}
    />
  </ModalStory>
)

export const Submitting = () => (
  <ModalStory>
    <SendStackOutputsModal
      phoneHomeId="ph-123"
      versionId="ver-abc"
      onSend={() => {}}
      isPending={true}
      error={undefined}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <SendStackOutputsModal
      phoneHomeId="ph-123"
      versionId="ver-abc"
      onSend={() => {}}
      isPending={false}
      error={{ error: 'Phone home endpoint not found', status: 404, description: '', user_error: false }}
    />
  </ModalStory>
)
