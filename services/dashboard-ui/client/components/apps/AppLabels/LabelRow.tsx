import { Badge } from '@/components/common/Badge'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import type { TAppLabelKey } from '@/lib/ctl-api/apps/get-app-labels'
import { LabelColorPicker } from './LabelColorPicker'

interface ILabelRow {
  label: TAppLabelKey
  disabled?: boolean
  onOverride: (key: string, color: string) => void
  onRemoveOverride: (key: string) => void
}

export const LabelRow = ({
  label,
  disabled,
  onOverride,
  onRemoveOverride,
}: ILabelRow) => {
  const entityTypes = [...(label.entity_types ?? [])].sort()
  const values = [...(label.values ?? [])].sort()

  return (
    <div className="flex flex-col gap-2 py-4 first:pt-0 last:pb-0">
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3 min-w-0 flex-wrap">
          <LabelBadge
            labelKey={label.key}
            labelValue={values[0] ?? ''}
            customColor={label.color}
            size="sm"
          />
          <span className="flex flex-wrap gap-1">
            {entityTypes.map((et) => (
              <Badge key={et} size="sm" theme="neutral">
                {et}
              </Badge>
            ))}
          </span>
          <Text variant="subtext" theme="neutral">
            {label.usage_count} use{label.usage_count !== 1 ? 's' : ''}
          </Text>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          {label.is_override && (
            <Badge size="sm" theme="brand">
              Custom
            </Badge>
          )}
          <LabelColorPicker
            id={`label-color-picker-${label.key}`}
            value={label.color}
            defaultColor={label.default_color}
            isOverride={label.is_override}
            disabled={disabled}
            ariaLabel={`Change color for ${label.key}`}
            onSelect={(color) => onOverride(label.key, color)}
            onReset={() => onRemoveOverride(label.key)}
          />
        </div>
      </div>

      {values.length > 1 && (
        <div className="flex items-center gap-2">
          <Text variant="subtext" theme="neutral">
            Values
          </Text>
          <span className="flex flex-wrap gap-1">
            {values.map((v) => (
              <Badge key={v} size="sm" theme="default">
                {v}
              </Badge>
            ))}
          </span>
        </div>
      )}
    </div>
  )
}
