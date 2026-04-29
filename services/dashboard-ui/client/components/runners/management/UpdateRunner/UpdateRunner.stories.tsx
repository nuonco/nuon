export default {
  title: 'Runners/UpdateRunner',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { UpdateRunnerModal } from './UpdateRunner'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <UpdateRunnerModal
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const UpdateManagerVersion = () => (
  <ModalStory>
    <UpdateRunnerModal
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
      modalHeading="Update manager version"
      inputLabel="Enter the manager version you'd like to update to."
      inputPlaceholder="manager version"
      submitLabel="Update manager version"
    />
  </ModalStory>
)

export const UpdateInstanceVersion = () => (
  <ModalStory>
    <UpdateRunnerModal
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
      modalHeading="Update instance version"
      inputLabel="Enter the instance version you'd like to update to."
      inputPlaceholder="instance version"
      submitLabel="Update instance version"
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <UpdateRunnerModal
      isPending={true}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <UpdateRunnerModal
      isPending={false}
      error={{ error: 'Unable to update runner.' } as any}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)
