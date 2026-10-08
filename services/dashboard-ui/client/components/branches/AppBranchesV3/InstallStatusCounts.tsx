import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TRolloutInstall } from './fixtures'
import { installStatusKey } from './InstallStatusTiles'

const STATUS_COUNTS = [
  { status: 'success', label: 'success' },
  { status: 'in-progress', label: 'in progress' },
  { status: 'approval-awaiting', label: 'awaiting approval' },
  { status: 'error', label: 'failed' },
  { status: 'cancelled', label: 'cancelled' },
  { status: 'pending', label: 'pending' },
]

export const InstallStatusCounts = ({
  installs,
  showTotal = true,
}: {
  installs: TRolloutInstall[]
  showTotal?: boolean
}) => {
  const counts = new Map<string, number>()
  for (const install of installs) {
    const key = installStatusKey(install)
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }
  const total = installs.length

  return (
    <>
      {showTotal ? (
        <Text variant="subtext" theme="neutral">
          {total} {total === 1 ? 'install' : 'installs'}
        </Text>
      ) : null}
      {STATUS_COUNTS.filter((item) => counts.get(item.status)).map((item) => (
        <Status key={item.status} status={item.status}>
          <span className="font-normal">
            {counts.get(item.status)} {item.label}
          </span>
        </Status>
      ))}
    </>
  )
}
