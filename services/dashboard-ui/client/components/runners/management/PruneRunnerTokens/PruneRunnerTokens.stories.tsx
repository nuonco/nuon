export default {
  title: 'Runners/PruneRunnerTokens',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { PruneRunnerTokensModal } from './PruneRunnerTokens'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <PruneRunnerTokensModal isLoading={false} onConfirm={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <PruneRunnerTokensModal isLoading={true} onConfirm={noop} />
  </ModalStory>
)
