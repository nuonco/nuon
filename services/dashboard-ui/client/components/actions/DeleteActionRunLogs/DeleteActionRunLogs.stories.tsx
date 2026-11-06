export default {
  title: 'Features / Actions / Delete action run logs',
}

import { ModalStory } from '@/components/__stories__/helpers'
import {
  ActionRunLogsDeleted,
  DeleteActionRunLogsButton,
  DeleteActionRunLogsModal,
} from './DeleteActionRunLogs'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DeleteActionRunLogsModal
      actionName="db-export"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <DeleteActionRunLogsModal
      actionName="db-export"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const AlreadyDeleted = () => (
  <ModalStory>
    <DeleteActionRunLogsModal
      actionName="db-export"
      isPending={false}
      error={null}
      logsDeleted
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DeleteActionRunLogsModal
      actionName="db-export"
      isPending={false}
      error={{
        error: 'action run is still running',
        description: 'logs can only be deleted after the action run finishes',
        user_error: true,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Deleted = () => <ActionRunLogsDeleted />

export const Button = () => <DeleteActionRunLogsButton onOpenModal={noop} />

export const ButtonWhileRunning = () => (
  <DeleteActionRunLogsButton
    disabled
    tooltipProps={{ tipContent: 'Logs can be deleted once the run finishes' }}
    onOpenModal={noop}
  />
)
