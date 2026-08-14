export default {
  title: 'Builds/CancelBuildModal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CancelBuildModal } from './CancelBuildModal'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CancelBuildModal componentName="api" onConfirm={noop} />
  </ModalStory>
)

export const NoComponentName = () => (
  <ModalStory>
    <CancelBuildModal onConfirm={noop} />
  </ModalStory>
)
