export default {
  title: 'Approvals/ApproveAll',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ApproveAllModal } from './ApproveAll'

const noop = () => {}

const mockSteps = [
  { id: 'step-1', name: 'deploy-terraform' },
  { id: 'step-2', name: 'deploy-helm-chart' },
  { id: 'step-3', name: 'apply-k8s-manifest' },
]

export const Default = () => (
  <ModalStory>
    <ApproveAllModal
      pendingSteps={mockSteps}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const SingleStep = () => (
  <ModalStory>
    <ApproveAllModal
      pendingSteps={[mockSteps[0]]}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ApproveAllModal
      pendingSteps={mockSteps}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ApproveAllModal
      pendingSteps={mockSteps}
      isPending={false}
      error={{ error: 'Some steps have already been responded to' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
