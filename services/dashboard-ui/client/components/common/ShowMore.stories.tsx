export default {
  title: 'UI / Show more',
}

import { useState } from 'react'
import { Text } from './Text'
import { ShowMore } from './ShowMore'

const SHORT = 'feat: add cache component'

const LONG = [
  'feat: map preview install configs onto branch groups',
  '',
  '* rename the first config to Customer A',
  '* rename the second config to Customer B',
  '* add Customer C install config',
  '* add Customer D install config',
  '* move prod install configs under install-configs/prod',
  '* pin AWS install configs to their app branches',
].join('\n')

const TextBlock = ({ text }: { text: string }) => (
  <Text as="p" variant="subtext" className="break-words whitespace-pre-line">
    {text}
  </Text>
)

export const Short = () => (
  <div className="max-w-md">
    <ShowMore>
      <TextBlock text={SHORT} />
    </ShowMore>
  </div>
)

export const Long = () => (
  <div className="max-w-md">
    <ShowMore>
      <TextBlock text={LONG} />
    </ShowMore>
  </div>
)

export const MaxHeight = () => (
  <div className="max-w-md">
    <ShowMore maxHeight={40}>
      <TextBlock text={LONG} />
    </ShowMore>
  </div>
)

export const PlusMinusIcons = () => (
  <div className="max-w-md">
    <ShowMore icons="plusMinus">
      <TextBlock text={LONG} />
    </ShowMore>
  </div>
)

export const Controlled = () => {
  const [expanded, setExpanded] = useState(false)

  return (
    <div className="flex flex-col gap-2 max-w-md">
      <ShowMore expanded={expanded} onExpandedChange={setExpanded}>
        <TextBlock text={LONG} />
      </ShowMore>
      <Text variant="subtext" theme="neutral">
        Controlled from outside: {expanded ? 'expanded' : 'collapsed'}
      </Text>
    </div>
  )
}
