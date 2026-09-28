import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import type { TInstallStatus, TInstallStatusAxis } from '@/types'

export const INSTALL_STATUS_AXES: {
  key: keyof TInstallStatus
  label: string
}[] = [
  { key: 'deployments', label: 'Deployments' },
  { key: 'resources', label: 'Resources' },
  { key: 'health_checks', label: 'Health checks' },
]

export interface IInstallStatusSummary {
  loading?: boolean
  status?: TInstallStatus
  unavailable?: boolean
}

export const InstallStatusAxis = ({
  axis,
  loading,
  unavailable,
}: {
  axis?: TInstallStatusAxis
  loading?: boolean
  unavailable?: boolean
}) => (
  <Status
    variant="badge"
    loading={loading}
    status={unavailable ? 'error' : axis?.status}
  >
    {unavailable ? 'Unavailable' : axis?.status_human_description}
  </Status>
)

export const InstallStatusSummary = ({
  loading,
  status,
  unavailable,
}: IInstallStatusSummary) => (
  <div className="flex items-start gap-6 shrink-0" aria-label="Install status">
    {INSTALL_STATUS_AXES.map(({ key, label }) => (
      <LabeledValue key={key} label={label}>
        <InstallStatusAxis
          axis={status?.[key]}
          loading={loading}
          unavailable={unavailable}
        />
      </LabeledValue>
    ))}
  </div>
)
