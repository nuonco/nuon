import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Time } from '@/components/common/Time'
import type { TTerraformWorkspaceLock } from '@/types'

export interface ITerraformWorkspaceLockBadge {
  lock: TTerraformWorkspaceLock
}

export const TerraformWorkspaceLockBadge = ({
  lock,
}: ITerraformWorkspaceLockBadge) => {
  const operation = lock.lock?.operation || 'Unknown operation'
  const who = lock.lock?.who
  const created = lock.lock?.created || lock.created_at

  return (
    <Badge variant="code" size="sm" theme="warn" className="flex items-center gap-1.5">
      <Icon variant="Lock" size={12} />
      <span>
        Locked by {operation}
        {who ? ` (${who})` : ''}
        {created && (
          <>
            {' — '}
            <Time time={created} format="relative" variant="label" />
          </>
        )}
      </span>
    </Badge>
  )
}
