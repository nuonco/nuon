import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useSurfaces } from '@/hooks/use-surfaces'
import { cn } from '@/utils/classnames'
import { RESOURCE_CATEGORIES, isCategoryOn } from './categories'
import { InterestsModal } from './InterestsModal'
import { ALL_RESOURCES, type Interests } from './types'

type Summary = { text: string; tone: 'neutral' | 'warn' }

const buildSummary = (value: Interests): Summary => {
  if (value.all_events) return { text: 'All events', tone: 'neutral' }

  const resources = value.resources ?? {}
  let count = 0
  for (const kind of ALL_RESOURCES) {
    const cfg = resources[kind]
    if (!cfg) continue
    for (const cat of RESOURCE_CATEGORIES[kind]) {
      if (isCategoryOn(cfg, cat)) count++
    }
  }

  if (count === 0) return { text: 'No events selected', tone: 'warn' }
  return {
    text: `${count} event${count === 1 ? '' : 's'} selected`,
    tone: 'neutral',
  }
}

// Compact summary + button. Opens InterestsModal in a stacked modal layer so
// the parent form (e.g. CreateWebhookModal) stays mounted underneath. The
// modal owns the draft and only commits back through onChange on Save.
export const InterestsPicker = ({
  value,
  onChange,
  disabled,
}: {
  value: Interests
  onChange: (next: Interests) => void
  disabled?: boolean
}) => {
  const { addModal } = useSurfaces()
  const summary = buildSummary(value)

  const openModal = () => {
    addModal(<InterestsModal value={value} onSave={onChange} />)
  }

  return (
    <Button
      type="button"
      variant="secondary"
      onClick={openModal}
      disabled={disabled}
      className="!justify-between !w-full"
    >
      <Text
        variant="body"
        theme={summary.tone === 'warn' ? 'warn' : 'default'}
        className={cn(summary.tone === 'warn' && 'font-strong')}
      >
        {summary.text}
      </Text>
      <span className="flex items-center gap-1">
        <Text variant="subtext" theme="neutral">
          Edit
        </Text>
        <Icon variant="PencilSimpleIcon" size={14} />
      </span>
    </Button>
  )
}
