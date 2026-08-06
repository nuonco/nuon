import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Tooltip } from '@/components/common/Tooltip'

export const AirgapBadge = ({ size = 'sm' }: { size?: 'sm' | 'md' | 'lg' }) => (
  <Tooltip
    tipContent="This install runs air-gapped from a bundle. Its live status is not reported back to Nuon."
    tipContentClassName="max-w-64 !whitespace-normal"
  >
    <Badge size={size} theme="info">
      <Icon variant="CloudSlashIcon" size={12} />
      Air-gapped
    </Badge>
  </Tooltip>
)
