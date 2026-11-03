export default {
  title: 'Features / Approvals / Deny plan',
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

export const DenyAndContinue = () => (
  <ModalStory>
    <DenyPlanModal
      modalCopy={{
        ...DENY_MODAL_COPY.terraform_plan,
        message:
          'This marks the plan as skipped and continues the workflow. If a later step in this group applies changes regardless of this plan, it will still run — check the workflow after skipping.',
      }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)
