import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useSurfaces } from '@/hooks/use-surfaces'
import { cn } from '@/utils/classnames'
import { describeMatch } from '@/components/match/types'
import type { SubscriptionMatch } from '@/components/match/types'
import { RESOURCE_CATEGORIES, isCategoryOn } from './categories'
import { InterestsModal, type PresetModalOutput } from './InterestsModal'
import { PRESETS, matchPreset } from './presets'
import { ALL_RESOURCES, type Interests } from './types'

type Summary = { text: string; tone: 'neutral' | 'warn' }

const buildSummary = (
  value: Interests,
  match: SubscriptionMatch | undefined
): Summary => {
  const presetId = matchPreset(value)
  if (presetId !== 'custom') {
    const preset = PRESETS.find((p) => p.id === presetId)
    if (preset) {
      const scopeText = match ? ` \u2014 ${describeMatch(match)}` : ''
      return { text: `${preset.label}${scopeText}`, tone: 'neutral' }
    }
  }

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
  const scopeText = match ? ` \u2014 ${describeMatch(match)}` : ''
  return {
    text: `${count} event${count === 1 ? '' : 's'} selected${scopeText}`,
    tone: 'neutral',
  }
}

export const InterestsPicker = ({
  value,
  matchValue,
  onChange,
  disabled,
}: {
  value: Interests
  matchValue?: SubscriptionMatch
  onChange: (output: PresetModalOutput) => void
  disabled?: boolean
}) => {
  const { addModal } = useSurfaces()
  const summary = buildSummary(value, matchValue)

  const openModal = () => {
    addModal(
      <InterestsModal
        value={value}
        matchValue={matchValue}
        onSave={onChange}
      />
    )
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
