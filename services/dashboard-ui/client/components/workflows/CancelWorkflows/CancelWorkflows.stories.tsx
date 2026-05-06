export default {
  title: 'Workflows/CancelWorkflows',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CancelWorkflowsModal } from './CancelWorkflows'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CancelWorkflowsModal
      count={3}
      isPending={false}
      error={null}
      cancelResults={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const SingleWorkflow = () => (
  <ModalStory>
    <CancelWorkflowsModal
      count={1}
      isPending={false}
      error={null}
      cancelResults={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <CancelWorkflowsModal
      count={3}
      isPending={true}
      error={null}
      cancelResults={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CancelWorkflowsModal
      count={3}
      isPending={false}
      error="Failed to connect to the server"
      cancelResults={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const PartialFailure = () => (
  <ModalStory>
    <CancelWorkflowsModal
      count={3}
      isPending={false}
      error={null}
      cancelResults={{
        cancelled: ['workflow-1', 'workflow-2'],
        errors: [
          {
            workflow_id: 'workflow-3',
            error: 'Workflow has already completed',
          },
        ],
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
