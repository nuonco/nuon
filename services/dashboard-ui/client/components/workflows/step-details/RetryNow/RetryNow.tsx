import { DateTime } from 'luxon'
import { useEffect, useState } from 'react'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

interface IRetryNowButton extends Omit<IButtonAsButton, 'onClick'> {
  retryNotBeforeAt?: string
  isPending: boolean
  onTrigger: () => void
}

const formatRemaining = (notBefore: DateTime, now: DateTime): string => {
  const seconds = Math.max(0, Math.round(notBefore.diff(now, 'seconds').seconds))
  if (seconds <= 0) return 'now'
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  if (m === 0) return `${s}s`
  if (s === 0) return `${m}m`
  return `${m}m ${s}s`
}

// RetryNowButton renders a live countdown until the step's RetryNotBeforeAt and
// fires retry-now on click to short-circuit the wait. Once the countdown
// reaches zero the executeworkflowstep handler will pick the step up on its
// next tick — at that point the button label says "now" until the polling
// query re-renders the step into in-progress.
export const RetryNowButton = ({
  retryNotBeforeAt,
  isPending,
  onTrigger,
  ...props
}: IRetryNowButton) => {
  const [now, setNow] = useState(DateTime.now())

  useEffect(() => {
    const id = setInterval(() => setNow(DateTime.now()), 1000)
    return () => clearInterval(id)
  }, [])

  if (!retryNotBeforeAt) return null

  const notBefore = DateTime.fromISO(retryNotBeforeAt)
  const remaining = formatRemaining(notBefore, now)

  return (
    <Button {...props} onClick={() => onTrigger()} disabled={isPending}>
      {isPending ? (
        <span className="flex items-center gap-2">
          <Icon variant="Loading" /> Retrying
        </span>
      ) : remaining === 'now' ? (
        'Retry now'
      ) : (
        <span className="flex items-center gap-2">
          <Icon variant="ClockIcon" />
          Retrying in {remaining} — retry now
        </span>
      )}
    </Button>
  )
}
