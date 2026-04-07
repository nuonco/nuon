export default {
  title: 'Runners/RunnerJobPlan',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RunnerJobPlanModal } from './RunnerJobPlan'

const mockPlan = {
  plan: {
    resources: [
      { type: 'aws_instance', name: 'web', action: 'create' },
      { type: 'aws_security_group', name: 'web_sg', action: 'create' },
    ],
  },
}

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <RunnerJobPlanModal
      plan={mockPlan as any}
      isLoading={false}
      error={null}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <RunnerJobPlanModal
      plan={undefined}
      isLoading={true}
      error={null}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RunnerJobPlanModal
      plan={undefined}
      isLoading={false}
      error={{ error: 'Unable to load runner job plan.' } as any}
    />
  </ModalStory>
)

export const NoPlan = () => (
  <ModalStory>
    <RunnerJobPlanModal
      plan={undefined}
      isLoading={false}
      error={null}
    />
  </ModalStory>
)
