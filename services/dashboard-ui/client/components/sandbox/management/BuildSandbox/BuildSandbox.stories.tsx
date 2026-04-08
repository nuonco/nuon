export default {
  title: 'Sandbox/BuildSandbox',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { BuildSandboxModal } from './BuildSandbox'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <BuildSandboxModal appName="My App" isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <BuildSandboxModal appName="My App" isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <BuildSandboxModal appName="My App" isPending={false} error={{ error: 'Build failed' }} onSubmit={noop} />
  </ModalStory>
)
