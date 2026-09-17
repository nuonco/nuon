import { cn } from '@/utils/classnames'
import type { TLabelSelector } from '../../utils/label-selector'
import { Badge } from '../atoms/Badge'
import { Text } from '../atoms/Text'

export interface ILabelSelectorSummary {
  selector?: TLabelSelector | null
  labelColors?: Record<string, string>
  loading?: boolean
  loadingWidth?: number
  className?: string
}

const ANY_VALUE = '*'

const sortedEntries = (labels?: Record<string, string>) =>
  Object.entries(labels ?? {}).sort(([a], [b]) => a.localeCompare(b))

const LabelChip = ({
  labelKey,
  labelValue,
  color,
  negated,
}: {
  labelKey: string
  labelValue: string
  color?: string
  negated?: boolean
}) => {
  const wildcard = labelValue === ANY_VALUE
  const chip = wildcard ? (
    <span className="inline-flex items-center gap-1">
      <Badge variant="code">{labelKey}</Badge>
      <Text variant="caption" color="tertiary">
        any
      </Text>
    </span>
  ) : (
    <Badge
      variant="code"
      labelKey={labelKey}
      labelValue={labelValue}
      color={color}
    />
  )

  if (!negated) return chip

  const name = wildcard ? `not ${labelKey} any` : `not ${labelKey}=${labelValue}`

  return (
    <span className="inline-flex items-center gap-1" aria-label={name}>
      <Text variant="label" className="text-status-error">
        not
      </Text>
      {chip}
    </span>
  )
}

export const LabelSelectorSummary = ({
  selector,
  labelColors,
  loading = false,
  loadingWidth,
  className,
}: ILabelSelectorSummary) => {
  if (loading) {
    return (
      <div className={cn('flex flex-wrap items-center gap-1.5', className)}>
        <Badge loading loadingWidth={loadingWidth ?? 10} variant="code" />
        <Badge loading loadingWidth={8} variant="code" />
      </div>
    )
  }

  const match = sortedEntries(selector?.match_labels)
  const notMatch = sortedEntries(selector?.not_match_labels)

  if (match.length === 0 && notMatch.length === 0) return null

  return (
    <div className={cn('flex flex-wrap items-center gap-1.5', className)}>
      {match.map(([key, value]) => (
        <LabelChip
          key={`match:${key}`}
          labelKey={key}
          labelValue={value}
          color={labelColors?.[key]}
        />
      ))}
      {notMatch.map(([key, value]) => (
        <LabelChip
          key={`not:${key}`}
          labelKey={key}
          labelValue={value}
          color={labelColors?.[key]}
          negated
        />
      ))}
    </div>
  )
}
