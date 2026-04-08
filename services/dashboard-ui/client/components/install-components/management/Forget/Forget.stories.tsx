export default {
  title: 'Install Components/Forget',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ForgetComponentModal } from './Forget'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <ForgetComponentModal componentName="web-server" isLoading={false} error={null} onConfirm={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ForgetComponentModal componentName="web-server" isLoading={true} error={null} onConfirm={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ForgetComponentModal componentName="web-server" isLoading={false} error={{ error: 'Component still has active deploys' }} onConfirm={noop} />
  </ModalStory>
)
