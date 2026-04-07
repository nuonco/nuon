export default {
  title: 'Workflows/CancelWorkflow',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CancelWorkflowModal } from './CancelWorkflow'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CancelWorkflowModal
      workflowType="deploy"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <CancelWorkflowModal
      workflowType="deploy"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CancelWorkflowModal
      workflowType="deploy"
      isPending={false}
      error={{ error: 'Workflow has already completed' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
