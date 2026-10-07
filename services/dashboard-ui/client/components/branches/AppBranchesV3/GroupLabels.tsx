import { Badge } from '@/components/common/Badge'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import type { TRolloutInstallGroup } from './fixtures'

const Label = ({ labelKey, value }: { labelKey: string; value: string }) => (
  <LabelBadge
    labelKey={labelKey}
    labelValue={value}
    size="sm"
    className="min-w-0 max-w-full [&>span]:min-w-0 [&_.block]:max-w-[16rem]"
  />
)

export const GroupLabels = ({
  group,
  max,
}: {
  group: TRolloutInstallGroup
  max?: number
}) => {
  const labels = Object.entries(group.label_selector?.match_labels ?? {})

  if (!labels.length) {
    return (
      <Text variant="subtext" theme="neutral">
        {group.default ? 'Every other install' : 'No label selector'}
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
