import type { ReactNode } from 'react'
import { CompositeError } from '@/components/common/CompositeError'
import { Link } from '@/components/common/Link'
import type { TCompositeError } from '@/types'
import type { TOverviewStage, TPreviewProgress } from './overview-loading'
import type { TGroupApprovedVersion } from './group-approved-version'
import type { IGroupPlanApproval } from './RolloutGroupsCard'
import type { TTrackGroup, TTrackInstall } from './RolloutTrack'
import type { TChangedBuildRow } from './changed-builds'
import { type IRunSourceCard } from './RunSourceCard'

export const InstallFailureNotice = ({
  error,
  href,
}: {
  error: TCompositeError
  href?: string
}) => (
  <div className="flex flex-col gap-2">
    <CompositeError error={error} />
    {href ? <Link href={href}>View install workflow</Link> : null}
  </div>
)

export interface TOverviewRollout extends IRunSourceCard {
  id: string
  href: string
  activity?: string
}

export interface TFailedBuildLink {
  id: string
  name: string
  href: string
}

export interface IBranchOverview {
  hasPlan: boolean
  showInstalls?: boolean
  showRolloutLink?: boolean
  previewMode?: TPreviewProgress
  isLoading?: boolean
  rollout?: TOverviewRollout
  changes?: ReactNode
  configChanges?: ReactNode
  // Changes when configChanges content changes, so open drawers can refresh.
  configChangesKey?: string
  groups: TTrackGroup[]
  rolloutHref: string
  groupHref: (groupId: string) => string
  loadingStages?: TOverviewStage[]
  buildRows?: TChangedBuildRow[]
  buildRowsLoading?: boolean
  compositeError?: TCompositeError
  failedBuilds?: TFailedBuildLink[]
  approvals?: IGroupPlanApproval[]
  runHeaderAction?: ReactNode
  cancelRollout?: ReactNode
  versionLabel?: string
  previousSha?: string
  onSelectInstall?: (install: TTrackInstall) => void
  onSelectBuild?: (build: TChangedBuildRow) => void
  onApproveAll?: (groupId: string) => void
  approvedVersions?: Record<string, TGroupApprovedVersion>
  approvedVersionsLoading?: boolean
}
