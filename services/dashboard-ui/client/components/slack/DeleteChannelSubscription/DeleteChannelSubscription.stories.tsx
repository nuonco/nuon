import { ModalStory } from '@/components/__stories__/helpers'
import { DeleteChannelSubscriptionModal } from './DeleteChannelSubscription'

export default { title: 'Features / Slack / Delete channel subscription' }

export const Default = () => (
  <ModalStory>
    <DeleteChannelSubscriptionModal
      channelLabel="#deploys"
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)
