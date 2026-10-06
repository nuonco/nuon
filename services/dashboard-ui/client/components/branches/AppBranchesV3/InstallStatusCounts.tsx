import { Status } from '@/components/common/Status'
import type { TRolloutInstall, TRolloutInstallGroup } from './fixtures'

const STATUS_COUNTS = [
  { status: 'success', label: 'success' },
  { status: 'in-progress', label: 'in progress' },
  { status: 'approval-awaiting', label: 'awaiting approval' },
  { status: 'error', label: 'failed' },
  { status: 'pending', label: 'pending' },
]

const countKey = (install: TRolloutInstall) =>
  install.awaitingApproval ? 'approval-awaiting' : install.status

export const InstallStatusCounts = ({
  group,
}: {
  group: TRolloutInstallGroup
}) => {
  const counts = new Map<string, number>()
  for (const install of group.installs) {
    const key = countKey(install)
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }

  return STATUS_COUNTS.filter((item) => counts.get(item.status)).map((item) => (
    <Status key={item.status} status={item.status}>
      <span className="font-normal">
        {counts.get(item.status)} {item.label}
      </span>
    </Status>
  ))
}
