export default {
  title: 'Installs/Forget',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ForgetModal } from './Forget'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <ForgetModal installName="prod-acme" isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ForgetModal installName="prod-acme" isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ForgetModal installName="prod-acme" isPending={false} error={{ error: 'Something went wrong' }} onSubmit={noop} />
  </ModalStory>
)
