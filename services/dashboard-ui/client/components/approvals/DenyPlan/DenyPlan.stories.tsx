export default {
  title: 'Approvals/DenyPlan',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DENY_MODAL_COPY } from '@/utils/approval-utils'
import { DenyPlanModal } from './DenyPlan'

const noop = () => {}

export const TerraformPlan = () => (
  <ModalStory>
    <DenyPlanModal
      modalCopy={DENY_MODAL_COPY.terraform_plan}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const HelmPlan = () => (
  <ModalStory>
    <DenyPlanModal
      modalCopy={DENY_MODAL_COPY.helm_approval}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <DenyPlanModal
      modalCopy={DENY_MODAL_COPY.terraform_plan}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DenyPlanModal
      modalCopy={DENY_MODAL_COPY.terraform_plan}
      isPending={false}
      error={{ error: 'Cannot deny while operation is in progress' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
