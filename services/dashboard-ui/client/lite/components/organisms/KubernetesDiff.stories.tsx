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
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { KubernetesDiff } from './KubernetesDiff'

export default {
  title: 'lite/organisms/KubernetesDiff',
}

const Frame = ({
  plan,
}: {
  plan: Parameters<typeof KubernetesDiff>[0]['plan']
}) => (
  <div className="max-w-6xl p-8">
    <KubernetesDiff plan={plan} />
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="KubernetesDiff"
    tier="organism"
    summary="A Kubernetes plan normalized into the shared summary, filter and diff sections."
    use={[
      'Review every Kubernetes resource changed by one apply operation.',
      'Search resource identity and filter create, update or delete operations.',
    ]}
    avoid={[
      'Do not pass pre-rendered unified diff lines.',
      'Do not put provider-specific controls inside individual sections.',
    ]}
    rules={[
      'Entry arrays normalize into complete before and after YAML.',
      'Planner errors remain visible as resource sections.',
      'All legacy Kubernetes diff scenarios have a matching lite story.',
    ]}
    props={[
      {
        name: 'plan',
        type: 'TKubernetesPlan',
        description: 'Raw Kubernetes planner response.',
      },
      {
        name: 'defaultOpen',
        type: 'boolean',
        default: 'true',
        description: 'Starting disclosure state for every resource.',
      },
    ]}
  />
)

export const Default = () => <Frame plan={mockPlan} />

export const DeploymentUpgrade = () => <Frame plan={deploymentUpgradePlan} />

export const FreshInstall = () => <Frame plan={freshInstallPlan} />

export const DeleteResources = () => <Frame plan={deleteResourcesPlan} />

export const MixedOperations = () => <Frame plan={mixedOperationsPlan} />

export const WithErrors = () => <Frame plan={withErrorsPlan} />

export const LongConfigMapEntries = () => (
  <Frame plan={longConfigMapEntriesPlan} />
)

export const SingleImageTagChange = () => (
  <Frame plan={singleImageTagChangePlan} />
)

export const LargeManifestSingleChange = () => (
  <Frame plan={largeSingleChangePlan} />
)

export const LargeManifestScatteredChanges = () => (
  <Frame plan={largeScatteredChangesPlan} />
)
