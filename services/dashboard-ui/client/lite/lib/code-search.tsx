import type { KeyboardEvent } from 'react'
import { Icon } from '../components/atoms/Icon'
import { Kbd } from '../components/atoms/Kbd'
import { Text } from '../components/atoms/Text'

export const matchNavKeyDown =
  (matchIndex: number, goTo: (index: number) => void) =>
  (event: KeyboardEvent<HTMLInputElement>) => {
    const back =
      event.key === 'ArrowUp' || (event.key === 'Enter' && event.shiftKey)
    const forward = event.key === 'ArrowDown' || event.key === 'Enter'
    if (!back && !forward) return

    event.preventDefault()
    goTo(back ? matchIndex - 1 : matchIndex + 1)
  }

const tooltip = (label: string, key: 'ArrowUpIcon' | 'ArrowDownIcon') => (
  <span className="flex items-center gap-1.5">
    <Text variant="caption">{label}</Text>
    <Kbd>
      <Icon variant={key} size={10} />
    </Kbd>
  </span>
)

export const MATCH_NAV_TOOLTIP = {
  previous: tooltip('Previous match', 'ArrowUpIcon'),
  next: tooltip('Next match', 'ArrowDownIcon'),
}
