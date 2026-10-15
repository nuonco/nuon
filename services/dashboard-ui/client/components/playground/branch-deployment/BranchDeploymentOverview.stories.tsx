export default {
  title: 'Playground / Branches / Deployment overview',
  fullBleed: true,
}

import { BranchDeploymentOverview } from './BranchDeploymentOverview'
import {
  awaitingApprovalOverview,
  longPlanOverview,
  rollingOutOverview,
} from './fixtures'

export const RollingOut = () => (
  <BranchDeploymentOverview overview={rollingOutOverview} />
)
RollingOut.storyName = 'Rolling out'

export const ActiveGroup = () => (
  <BranchDeploymentOverview overview={rollingOutOverview} />
)
ActiveGroup.storyName = 'Active group'

export const GroupInstalls = () => (
  <BranchDeploymentOverview
    overview={rollingOutOverview}
    initialGroupId="d1"
    initialTab="installs"
  />
)
GroupInstalls.storyName = 'Group installs'

export const GroupChanges = () => (
  <BranchDeploymentOverview
    overview={rollingOutOverview}
    initialGroupId="d1"
    initialTab="changes"
  />
)
GroupChanges.storyName = 'Group changes'

export const ManyGroups = () => (
  <BranchDeploymentOverview overview={longPlanOverview} />
)
ManyGroups.storyName = 'Many groups'

export const AwaitingApproval = () => (
  <BranchDeploymentOverview overview={awaitingApprovalOverview} />
)
AwaitingApproval.storyName = 'Awaiting approval'
