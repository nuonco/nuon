import { Badge } from '@/components/common/Badge'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import type { TGroupMatch } from '@/components/branches/BranchOverview/RolloutTrack'

const Label = ({ labelKey, value }: { labelKey: string; value: string }) => (
  <LabelBadge
    labelKey={labelKey}
    labelValue={value}
    size="sm"
    className="min-w-0 max-w-full [&>span]:min-w-0 [&_.block]:max-w-[16rem]"
  />
)

export const GroupLabels = ({
  match,
  max,
}: {
  match?: TGroupMatch
  max?: number
}) => {
  const labels = Object.entries(match?.labels ?? {})

  if (!labels.length) {
    return (
      <Text variant="subtext" theme="neutral">
        {match?.kind === 'default'
          ? 'Every other install'
          : match?.kind === 'pinned'
            ? 'Pinned installs'
            : 'No label selector'}
      </Text>
    )
  }

  const visible = max ? labels.slice(0, max) : labels
  const hidden = labels.slice(visible.length)

  return (
    <span className="flex min-w-0 max-w-full flex-wrap items-center gap-1.5">
      {visible.map(([key, value]) => (
        <Label key={key} labelKey={key} value={value} />
      ))}
      {hidden.length ? (
        <Tooltip
          position="top"
          tipContentClassName="whitespace-normal max-w-md"
          tipContent={
            <span className="flex flex-wrap gap-1.5 p-1">
              {hidden.map(([key, value]) => (
                <Label key={key} labelKey={key} value={value} />
              ))}
            </span>
          }
        >
          <Badge size="sm" theme="neutral" variant="code">
            +{hidden.length} more
          </Badge>
        </Tooltip>
      ) : null}
    </span>
  )
}
