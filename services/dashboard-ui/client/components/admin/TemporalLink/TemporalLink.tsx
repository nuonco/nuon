import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

interface ITemporalLink {
  href: string
  isVisible: boolean
}

export const TemporalLink = ({ href, isVisible }: ITemporalLink) => {
  return isVisible ? (
    <Button href={href} target="_blank">
      View in Temporal <Icon variant="ArrowSquareOutIcon" />
    </Button>
  ) : null
}
