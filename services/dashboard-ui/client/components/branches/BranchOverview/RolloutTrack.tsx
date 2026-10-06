import { useState, type ReactNode } from 'react'
import { Text } from '@/components/common/Text'
import {
  stepStatusCategory,
  type TStepStatusCategory,
} from '@/components/branches/shared/step-status'
import { RolloutGroupDetail } from './RolloutGroupDetail'
import { RolloutGroupsCard } from './RolloutGroupsCard'

export type TTrackStatus = {
  status: string
  detail?: string
}

export type TTrackInstall = {
  id: string
  name: string
  status: string
  detail?: string
  durationNs?: number
  href?: string
  resources?: TTrackStatus
  deployment?: TTrackStatus
  health?: TTrackStatus
  overviewHref?: string
  workflowId?: string
  workflowHref?: string
  labels?: Record<string, string>
  region?: string
  appliedConfigId?: string
}

export type TGroupMatch = {
  kind: 'labels' | 'default' | 'pinned'
  labels?: Record<string, string>
}

export type TTrackGroup = {
  id: string
  name: string
  status: string
  installs: TTrackInstall[]
  plannedCount?: number
  rules?: ReactNode
  match?: TGroupMatch
  approval?: string
  maxParallel?: number
}

export interface IRolloutTrack {
  groups: TTrackGroup[]
  initialGroupId?: string
  selectedGroupId?: string
  emptyMessage?: string
  summary?: ReactNode
  notice?: ReactNode
  footer?: ReactNode
  groupHref?: (groupId: string) => string
  onSelectGroup?: (groupId: string) => void
  onSelectInstall?: (groupId: string, installId: string) => void
}

export const selectedTrackGroup = (groups: TTrackGroup[], groupId?: string) =>
  groups.find((group) => group.id === groupId) ??
  groups.find((group) => group.id === defaultGroupId(groups))

const defaultGroupId = (groups: TTrackGroup[]) => {
  const byCategory = (category: TStepStatusCategory) =>
    groups.find((group) => stepStatusCategory(group.status) === category)
  return (
    byCategory('error') ??
    byCategory('awaiting') ??
    byCategory('active') ??
    (groups.every((group) => stepStatusCategory(group.status) === 'success')
      ? groups.at(-1)
      : groups[0])
  )?.id
}

export const RolloutTrack = ({
  groups,
  initialGroupId,
  selectedGroupId,
  emptyMessage,
  summary,
  notice,
  footer,
  groupHref,
  onSelectGroup,
  onSelectInstall,
}: IRolloutTrack) => {
  const [localId, setLocalId] = useState(initialGroupId)
  const fallbackId = defaultGroupId(groups)
  const chosenId = onSelectGroup || groupHref ? selectedGroupId : localId
  const selected =
    groups.find((group) => group.id === chosenId) ??
    groups.find((group) => group.id === fallbackId)
  const choose = (id: string) => {
    if (onSelectGroup) onSelectGroup(id)
    else setLocalId(id)
  }

  if (groups.length === 0) {
    return emptyMessage ? (
      <Text variant="subtext" theme="neutral">
        {emptyMessage}
      </Text>
    ) : null
  }

  return (
    <div className="flex flex-col gap-6">
      <RolloutGroupsCard
        groups={groups}
        selectedGroupId={selected?.id}
        groupHref={groupHref ?? ((id) => `#${id}`)}
        onSelectGroup={groupHref && !onSelectGroup ? undefined : choose}
      />
      {selected ? (
        <RolloutGroupDetail
          group={selected}
          notice={notice}
          footer={footer}
          summary={summary}
          onSelectInstall={onSelectInstall}
        />
      ) : null}
    </div>
  )
}
