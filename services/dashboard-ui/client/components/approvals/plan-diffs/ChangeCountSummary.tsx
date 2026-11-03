import { Text, type TTextTheme } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

interface IChangeCountSummary {
  added?: number
  updated?: number
  removed?: number
  replaced?: number
  emptyText?: string
  className?: string
}

export const ChangeCountSummary = ({
  added = 0,
  updated = 0,
  removed = 0,
  replaced = 0,
  emptyText = 'no changes',
  className,
}: IChangeCountSummary) => {
  const parts: { text: string; theme: TTextTheme; className?: string }[] = []
  if (added > 0) parts.push({ text: `+${added}`, theme: 'success' })
  if (updated > 0) parts.push({ text: `~${updated}`, theme: 'warn' })
  if (removed > 0) parts.push({ text: `-${removed}`, theme: 'error' })
  if (replaced > 0)
    parts.push({
      text: `±${replaced}`,
      theme: 'default',
      className: 'text-fuchsia-600 dark:text-fuchsia-300',
    })

  if (parts.length === 0) {
    return (
      <Text variant="subtext" theme="neutral" className={className}>
        {emptyText}
      </Text>
    )
  }

  return (
    <div className={cn('flex items-center gap-2', className)}>
      {parts.map((part) => (
        <Text
          key={part.text}
          variant="subtext"
          theme={part.theme}
          weight="strong"
          family="mono"
          className={part.className}
        >
          {part.text}
        </Text>
      ))}
    </div>
  )
}
