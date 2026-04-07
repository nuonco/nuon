export default {
  title: 'Runners/ShutdownInstance',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ShutdownInstanceModal } from './ShutdownInstance'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <ShutdownInstanceModal error={null} isLoading={false} onConfirm={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ShutdownInstanceModal error={null} isLoading={true} onConfirm={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ShutdownInstanceModal error={{ error: 'Instance restart failed' }} isLoading={false} onConfirm={noop} />
  </ModalStory>
)
