export default {
  title: 'InstallComponents/RecoverHelmRelease',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RecoverHelmReleaseModal } from './RecoverHelmRelease'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <RecoverHelmReleaseModal
      componentName="api-server"
      status="pending-upgrade"
      isPending={false}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <RecoverHelmReleaseModal
      componentName="api-server"
      status="pending-upgrade"
      isPending
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RecoverHelmReleaseModal
      componentName="api-server"
      status="pending-upgrade"
      isPending={false}
      error={
        {
          error:
            'A job is currently running for api-server. Wait for it to finish, or cancel it, before recovering the Helm release.',
        } as any
      }
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)
