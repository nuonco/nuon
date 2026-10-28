import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TAPIError } from '@/types'
import type { TInstallShare } from '../types'
import { ShareInstallModal } from './ShareInstallModal'
import type { ShareInstallFormValues } from './schema'

export const ShareInstallButton = ({
  installName,
  shares,
  isPending = false,
  error = null,
  onSubmit,
  onRevoke,
  onResend,
  ...buttonProps
}: {
  installName?: string
  shares?: TInstallShare[]
  isPending?: boolean
  error?: TAPIError | null
  onSubmit?: (params: ShareInstallFormValues) => void
  onRevoke?: (shareId: string) => void
  onResend?: (shareId: string) => void
} & Omit<IButtonAsButton, 'onClick'>) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      onClick={() =>
        addModal(
          <ShareInstallModal
            isPending={isPending}
            error={error}
            installName={installName}
            shares={shares}
            onSubmit={onSubmit ?? (() => {})}
            onRevoke={onRevoke}
            onResend={onResend}
          />
        )
      }
      {...buttonProps}
    >
      <Icon variant="ShareIcon" />
      Share install
    </Button>
  )
}
