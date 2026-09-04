import {
  deleteResourcesPlan,
  deploymentUpgradePlan,
  freshInstallPlan,
  largeScatteredChangesPlan,
  largeSingleChangePlan,
  longConfigMapEntriesPlan,
  mixedOperationsPlan,
  mockPlan,
  singleImageTagChangePlan,
  withErrorsPlan,
} from '@/lib/fixtures/plan-diffs/kubernetes'
import { KubernetesDiff } from './KubernetesDiff'

export default {
  title: 'Approvals/PlanDiffs/KubernetesDiff',
}

export const Default = () => <KubernetesDiff plan={mockPlan} />

export const DeploymentUpgrade = () => (
  <KubernetesDiff plan={deploymentUpgradePlan} />
)

export const FreshInstall = () => <KubernetesDiff plan={freshInstallPlan} />

export const DeleteResources = () => (
  <KubernetesDiff plan={deleteResourcesPlan} />
)

export const MixedOperations = () => (
  <KubernetesDiff plan={mixedOperationsPlan} />
)

export const WithErrors = () => <KubernetesDiff plan={withErrorsPlan} />

export const LongConfigMapEntries = () => (
  <KubernetesDiff plan={longConfigMapEntriesPlan} />
)

export const SingleImageTagChange = () => (
  <KubernetesDiff plan={singleImageTagChangePlan} />
)

export const LargeManifestSingleChange = () => (
  <KubernetesDiff plan={largeSingleChangePlan} />
)

export const LargeManifestScatteredChanges = () => (
  <KubernetesDiff plan={largeScatteredChangesPlan} />
)
