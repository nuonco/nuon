export default {
  title: 'Features / Branches / Branch overview playground',
  fullBleed: true,
}

import { BranchOverviewPlayground } from './BranchOverviewPlayground'
import {
  awaitingApprovalFixture,
  failedFixture,
  idleFixture,
  noPlanFixture,
  rollingOutFixture,
  succeededFixture,
} from './fixtures'

export const RollingOut = () => (
  <BranchOverviewPlayground branch={rollingOutFixture} />
)
RollingOut.storyName = 'Rolling out'

export const WaitingForApproval = () => (
  <BranchOverviewPlayground branch={awaitingApprovalFixture} />
)
WaitingForApproval.storyName = 'Waiting for approval'

export const Succeeded = () => (
  <BranchOverviewPlayground branch={succeededFixture} />
)

export const Failed = () => <BranchOverviewPlayground branch={failedFixture} />

export const NoRunsYet = () => <BranchOverviewPlayground branch={idleFixture} />
NoRunsYet.storyName = 'No runs yet'

export const NoDeploymentPlan = () => (
  <BranchOverviewPlayground branch={noPlanFixture} />
)
NoDeploymentPlan.storyName = 'No deployment plan'

export const Rollout = () => (
  <BranchOverviewPlayground branch={rollingOutFixture} initialView="rollout" />
)

export const RolloutFailed = () => (
  <BranchOverviewPlayground branch={failedFixture} initialView="rollout" />
)
RolloutFailed.storyName = 'Rollout failed'

export const Settings = () => (
  <BranchOverviewPlayground branch={rollingOutFixture} initialView="settings" />
)

export const InspectTemplate = () => (
  <BranchOverviewPlayground branch={rollingOutFixture} initialView="template" />
)
InspectTemplate.storyName = 'Inspect template'
