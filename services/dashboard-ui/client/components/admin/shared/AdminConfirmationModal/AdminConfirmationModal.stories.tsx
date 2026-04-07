export default {
  title: 'Admin/AdminConfirmationModal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { AdminConfirmationModal } from './AdminConfirmationModal'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <AdminConfirmationModal
      title="Confirm action"
      message="Are you sure you want to proceed with this action?"
      onConfirm={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const Warning = () => (
  <ModalStory>
    <AdminConfirmationModal
      title="Restart runner"
      message="This will restart the org runner."
      variant="warning"
      onConfirm={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const Danger = () => (
  <ModalStory>
    <AdminConfirmationModal
      title="Force shutdown"
      message="This will forcefully shutdown the install runner and may cause data loss."
      variant="danger"
      onConfirm={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const RequiresInput = () => (
  <ModalStory>
    <AdminConfirmationModal
      title="Deprovision org"
      message="This will deprovision ALL infrastructure for this organization."
      variant="danger"
      requiresInput
      inputText="yesimsure"
      onConfirm={noop}
      onCancel={noop}
    />
  </ModalStory>
)
