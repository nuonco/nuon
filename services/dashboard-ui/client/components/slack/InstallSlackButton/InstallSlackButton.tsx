import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

export const InstallSlackButton = ({
  isPending,
  onInstall,
  ...props
}: {
  isPending: boolean
  onInstall: () => void
} & Omit<IButtonAsButton, 'children' | 'onClick'>) => (
  <Button
    variant="primary"
    onClick={() => onInstall()}
    disabled={isPending || props.disabled}
    {...props}
  >
    {isPending ? (
      <span className="flex items-center gap-2">
        <Icon variant="Loading" /> Redirecting…
      </span>
    ) : (
      <span className="flex items-center gap-2">
        <Icon variant="SlackLogoIcon" />
        Add to Slack
      </span>
    )}
  </Button>
)
