export default {
  title: 'Installs/Deprovision',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeprovisionModal } from './Deprovision'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DeprovisionModal installName="prod-acme" isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <DeprovisionModal installName="prod-acme" isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DeprovisionModal installName="prod-acme" isPending={false} error={{ error: 'Something went wrong' }} onSubmit={noop} />
  </ModalStory>
)
