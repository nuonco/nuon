export default {
  title: 'Sandbox/DriftScanSandbox',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DriftScanSandboxModal } from './DriftScanSandbox'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DriftScanSandboxModal isPending={false} error={null} onSubmit={noop} onClose={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <DriftScanSandboxModal isPending={true} error={null} onSubmit={noop} onClose={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DriftScanSandboxModal isPending={false} error={{ error: 'Drift scan failed' }} onSubmit={noop} onClose={noop} />
  </ModalStory>
)
