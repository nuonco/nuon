import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import type { TAppBranchInstallGroup } from '@/types'
import type { TGroupMatch } from './RolloutTrack'

export const installGroupMatch = (
  group?: TAppBranchInstallGroup
): TGroupMatch | undefined => {
  if (!group) return undefined
  const labels = group.label_selector?.match_labels
  if (labels && Object.keys(labels).length > 0) {
    return { kind: 'labels', labels }
  }
  if (group.default) return { kind: 'default' }
  return { kind: 'pinned' }
}

export const InstallGroupMatch = ({
  match,
  pace,
}: {
  match: TGroupMatch
  pace?: string
}) => (
  <div className="flex w-fit max-w-md flex-col gap-2 rounded-xl border bg-white px-4 py-3 shadow-sm dark:bg-dark-grey-900">
    {match.kind === 'labels' ? (
      <>
        <Text variant="subtext" theme="neutral">
          Matches labels
        </Text>
        <span className="flex flex-wrap gap-2">
          {Object.entries(match.labels ?? {}).map(([key, value]) => (
            <LabelBadge key={key} labelKey={key} labelValue={value} size="sm" />
          ))}
        </span>
      </>
    ) : match.kind === 'default' ? (
      <>
        <Text variant="subtext" weight="strong">
          Default group
        </Text>
        <Text variant="subtext" theme="neutral">
          Installs that do not match another group.
        </Text>
      </>
    ) : (
      <>
        <Text variant="subtext" weight="strong">
          Pinned installs
        </Text>
        <Text variant="subtext" theme="neutral">
          Only installs assigned to this group.
        </Text>
      </>
    )}
    {pace ? (
      <Text variant="subtext" theme="neutral">
        {pace}
      </Text>
    ) : null}
  </div>
)
