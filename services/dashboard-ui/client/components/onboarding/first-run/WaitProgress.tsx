import { useEffect, useState } from 'react'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { AWESOME_BYOC_URL, BYOC_REFERENCES } from './constants'
import { formatElapsed, type IWaitItem } from './progress'

const ROTATE_MS = 7000
const FADE_MS = 250

const Elapsed = ({ since }: { since: number }) => {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [])
  return (
    <Text variant="subtext" theme="neutral" className="tabular-nums">
      {formatElapsed(now - since)}
    </Text>
  )
}

const ItemIcon = ({ state }: { state: IWaitItem['state'] }) =>
  state === 'done' ? (
    <Icon variant="CheckCircleIcon" size={16} theme="success" weight="fill" />
  ) : state === 'active' ? (
    <Icon variant="Loading" size={16} />
  ) : state === 'error' ? (
    <Icon variant="WarningCircleIcon" size={16} theme="error" weight="fill" />
  ) : (
    <span className="mx-px inline-block size-3.5 rounded-full border" aria-hidden />
  )

export const WaitProgress = ({
  title,
  items,
  since,
}: {
  title: string
  items: IWaitItem[]
  since?: number
}) => (
  <Card className="!gap-4 !p-5">
    <div className="flex items-center justify-between gap-3">
      <Text variant="base" weight="strong">
        {title}
      </Text>
      {since ? <Elapsed since={since} /> : null}
    </div>
    <ol className="flex flex-col gap-2.5" aria-live="polite">
      {items.map((item) => (
        <li key={item.label} className="flex items-center gap-2.5">
          <ItemIcon state={item.state} />
          <Text
            variant="body"
            theme={item.state === 'pending' ? 'neutral' : item.state === 'error' ? 'error' : undefined}
            weight={item.state === 'active' ? 'strong' : undefined}
          >
            {item.label}
          </Text>
        </li>
      ))}
    </ol>
  </Card>
)

export const ByocReferences = ({ start = 0 }: { start?: number }) => {
  const [index, setIndex] = useState(start % BYOC_REFERENCES.length)
  const [visible, setVisible] = useState(true)
  const [paused, setPaused] = useState(false)

  useEffect(() => {
    if (paused) return
    let swap: ReturnType<typeof setTimeout> | undefined
    const timer = setInterval(() => {
      setVisible(false)
      swap = setTimeout(() => {
        setIndex((i) => (i + 1) % BYOC_REFERENCES.length)
        setVisible(true)
      }, FADE_MS)
    }, ROTATE_MS)
    return () => {
      clearInterval(timer)
      if (swap) clearTimeout(swap)
    }
  }, [paused])

  const ref = BYOC_REFERENCES[index]

  return (
    <Card className="!gap-3 !p-5" onMouseEnter={() => setPaused(true)} onMouseLeave={() => setPaused(false)}>
      <div className="flex items-center justify-between gap-3">
        <Text variant="subtext" theme="neutral">
          Shipping in their customers&apos; clouds
        </Text>
        <Link href={AWESOME_BYOC_URL} isExternal textVariant="subtext">
          See all on Awesome BYOC
        </Link>
      </div>
      <div
        className="flex flex-col gap-1 transition-opacity duration-200 motion-reduce:transition-none"
        style={{ opacity: visible ? 1 : 0 }}
      >
        <div className="flex flex-wrap items-baseline gap-x-2">
          <Link href={ref.href} isExternal textVariant="base">
            {ref.name}
          </Link>
          <Text variant="subtext" theme="neutral">
            {ref.category}
          </Text>
        </div>
        <Text variant="body" theme="neutral">
          {ref.blurb}
        </Text>
      </div>
    </Card>
  )
}
