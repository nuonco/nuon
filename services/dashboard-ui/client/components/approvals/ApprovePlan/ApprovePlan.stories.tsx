export default {
  title: 'Approvals/ApprovePlan',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { APPROVAL_MODAL_COPY } from '@/utils/approval-utils'
import { ApprovePlanModal } from './ApprovePlan'

const noop = () => {}

export const TerraformPlan = () => (
  <ModalStory>
    <ApprovePlanModal
      modalCopy={APPROVAL_MODAL_COPY.terraform_plan}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const HelmPlan = () => (
  <ModalStory>
    <ApprovePlanModal
      modalCopy={APPROVAL_MODAL_COPY.helm_approval}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ApprovePlanModal
      modalCopy={APPROVAL_MODAL_COPY.terraform_plan}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ApprovePlanModal
      modalCopy={APPROVAL_MODAL_COPY.terraform_plan}
      isPending={false}
      error={{ error: 'Install is currently busy with another operation' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
