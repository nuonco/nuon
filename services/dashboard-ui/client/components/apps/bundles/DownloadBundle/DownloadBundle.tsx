import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

interface IDownloadBundle extends IButtonAsButton {
  isPending: boolean
}

export const DownloadBundle = ({ isPending, ...props }: IDownloadBundle) => (
  <Button variant="ghost" size="sm" disabled={isPending} {...props}>
    <Icon variant={isPending ? 'Loading' : 'DownloadIcon'} size={16} />
    {isPending ? 'Preparing download' : 'Download'}
  </Button>
)
