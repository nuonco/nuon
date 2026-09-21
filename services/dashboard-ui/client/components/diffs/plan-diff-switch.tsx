import type {
  THelmPlan,
  TKubernetesPlan,
  TPulumiPlan,
  TTerraformPlan,
} from '@/types'
import { useDashboardPreferences } from '@/hooks/use-dashboard-preferences'
import { HelmDiff as LegacyHelmDiff } from '@/components/approvals/plan-diffs/helm/HelmDiff'
import { KubernetesDiff as LegacyKubernetesDiff } from '@/components/approvals/plan-diffs/kubernetes/KubernetesDiff'
import { PulumiDiff as LegacyPulumiDiff } from '@/components/approvals/plan-diffs/pulumi/PulumiDiff'
import { TerraformDiff as LegacyTerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { HelmDiff as HelmDiffV2 } from './HelmDiff'
import { KubernetesDiff as KubernetesDiffV2 } from './KubernetesDiff'
import { PulumiDiff as PulumiDiffV2 } from './PulumiDiff'
import { TerraformDiff as TerraformDiffV2 } from './TerraformDiff'

export const TerraformDiff = ({
  plan,
}: {
  plan: TTerraformPlan | undefined
}) => {
  const { diffViewer } = useDashboardPreferences()

  return diffViewer === 'v2' ? (
    <TerraformDiffV2 plan={plan} />
  ) : (
    <LegacyTerraformDiff plan={plan} />
  )
}

export const HelmDiff = ({ plan }: { plan: THelmPlan }) => {
  const { diffViewer } = useDashboardPreferences()

  return diffViewer === 'v2' ? (
    <HelmDiffV2 plan={plan} />
  ) : (
    <LegacyHelmDiff plan={plan} />
  )
}

export const KubernetesDiff = ({ plan }: { plan: TKubernetesPlan }) => {
  const { diffViewer } = useDashboardPreferences()

  return diffViewer === 'v2' ? (
    <KubernetesDiffV2 plan={plan} />
  ) : (
    <LegacyKubernetesDiff plan={plan} />
  )
}

type TLegacyPulumiPlan = Parameters<typeof LegacyPulumiDiff>[0]['plan']

export const PulumiDiff = ({ plan }: { plan: TLegacyPulumiPlan }) => {
  const { diffViewer } = useDashboardPreferences()

  return diffViewer === 'v2' ? (
    <PulumiDiffV2 plan={plan as TPulumiPlan | undefined} />
  ) : (
    <LegacyPulumiDiff plan={plan} />
  )
}
