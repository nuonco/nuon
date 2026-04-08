export default {
  title: 'Runners/ShutdownMngRunner',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ShutdownMngRunnerModal } from './ShutdownMngRunner'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <ShutdownMngRunnerModal label="Shutdown process" error={null} isLoading={false} onConfirm={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ShutdownMngRunnerModal label="Shutdown process" error={null} isLoading={true} onConfirm={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ShutdownMngRunnerModal label="Shutdown process" error={{ error: 'Shutdown failed' }} isLoading={false} onConfirm={noop} />
  </ModalStory>
)
