export default {
  title: 'Approvals/RetryPlan',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RETRY_MODAL_COPY } from '@/utils/approval-utils'
import { RetryPlanModal } from './RetryPlan'

const noop = () => {}

export const TerraformPlan = () => (
  <ModalStory>
    <RetryPlanModal
      modalCopy={RETRY_MODAL_COPY.terraform_plan}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const HelmPlan = () => (
  <ModalStory>
    <RetryPlanModal
      modalCopy={RETRY_MODAL_COPY.helm_approval}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <RetryPlanModal
      modalCopy={RETRY_MODAL_COPY.terraform_plan}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RetryPlanModal
      modalCopy={RETRY_MODAL_COPY.terraform_plan}
      isPending={false}
      error={{ error: 'Workflow has already completed' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
