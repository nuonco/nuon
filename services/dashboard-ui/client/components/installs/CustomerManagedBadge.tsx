import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Tooltip } from '@/components/common/Tooltip'

export const CustomerManagedBadge = ({
  size = 'sm',
}: {
  size?: 'sm' | 'md' | 'lg'
}) => (
  <Tooltip
    tipContent="This customer-managed install runs offline from a bundle. Its live status is not reported to Nuon."
    tipContentClassName="max-w-64 !whitespace-normal"
  >
    <Badge size={size} theme="info">
      <Icon variant="CloudSlashIcon" size={12} />
      Customer-managed · Offline
    </Badge>
  </Tooltip>
)
