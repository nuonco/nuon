export default {
  title: 'Installs/Reprovision',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ReprovisionModal } from './Reprovision'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <ReprovisionModal installName="prod-acme" isPending={false} error={null} onSubmit={noop} roleSelector={null} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ReprovisionModal installName="prod-acme" isPending={true} error={null} onSubmit={noop} roleSelector={null} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ReprovisionModal installName="prod-acme" isPending={false} error={{ error: 'Something went wrong' }} onSubmit={noop} roleSelector={null} />
  </ModalStory>
)
