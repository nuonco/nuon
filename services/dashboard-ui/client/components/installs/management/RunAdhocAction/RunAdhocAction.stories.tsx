export default {
  title: 'Installs/RunAdhocAction',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RunAdhocActionModal } from './RunAdhocAction'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <RunAdhocActionModal
      installId="inst-123"
      isPending={false}
      error={null}
      onSubmit={noop}
      onDraftResume={noop as any}
      roleSelector={null}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <RunAdhocActionModal
      installId="inst-123"
      isPending={true}
      error={null}
      onSubmit={noop}
      onDraftResume={noop as any}
      roleSelector={null}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RunAdhocActionModal
      installId="inst-123"
      isPending={false}
      error={{ error: 'Something went wrong' }}
      onSubmit={noop}
      onDraftResume={noop as any}
      roleSelector={null}
    />
  </ModalStory>
)

export const WithInitialValues = () => (
  <ModalStory>
    <RunAdhocActionModal
      installId="inst-123"
      initialValues={{ command: 'echo hello', name: 'test action', timeout: 600 }}
      isPending={false}
      error={null}
      onSubmit={noop}
      onDraftResume={noop as any}
      roleSelector={null}
    />
  </ModalStory>
)
