import { forwardRef } from 'react'
import { Button } from '@/components/common/Button'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'

const CONSENT_TIMEOUT = 1000 * 60 * 60 * 24

interface IConsentToast {
  onAccept: () => void
  onDecline: () => void
  toastId?: string
  pauseTimeout?: boolean
}

export const ConsentToast = forwardRef<HTMLDivElement, IConsentToast>(
  ({ onAccept, onDecline, toastId, pauseTimeout }, ref) => {
    const { removeToast } = useToast()

    const decide = (choose: () => void) => () => {
      choose()
      if (toastId) removeToast(toastId)
    }

    return (
      <Toast
        ref={ref}
        className="!w-106"
        toastId={toastId}
        pauseTimeout={pauseTimeout}
        heading="Help us improve Nuon"        
        theme="default"
        timeout={CONSENT_TIMEOUT}
      >
        <Text className="leading-normal" variant="subtext">
          We use product analytics to understand how the dashboard is used.
          Declining won't change how Nuon works.
        </Text>
        <span className="flex items-center gap-3 my-3">
          <Button variant="primary" onClick={decide(onAccept)} size="sm">
            Accept
          </Button>
          <Button variant="secondary" onClick={decide(onDecline)} size="sm">
            Decline
          </Button>
        </span>
        <Link
          href="https://www.nuon.co/privacy"
          isExternal
          className="inline-flex items-center gap-1"
        >
          Privacy policy
        </Link>
      </Toast>
    )
  }
)

ConsentToast.displayName = 'ConsentToast'
