import { Banner } from '@/components/common/Banner'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'

export type TInstallCronOfflineKind =
  | 'action'
  | 'sandbox_drift'
  | 'component_drift'

const COPY: Record<
  TInstallCronOfflineKind,
  { title: string; description: string }
> = {
  action: {
    title: 'Action crons are paused',
    description:
      'The runner is offline, cron schedulings action runs are disabled. They will resume automatically when the runner is back online.',
  },
  sandbox_drift: {
    title: 'Sandbox drift checks are paused',
    description:
      'The runner is offline, cron schedulings for sandbox drift runs are disabled. They will resume automatically when the runner is back online.',
  },
  component_drift: {
    title: 'Component drift checks are paused',
    description:
      'The runner is offline, cron schedulings for component drift runs are disabled. They will resume automatically when the runner is back online.',
  },
}

export const InstallCronOfflineBanner = ({
  runnerStatus,
  kind,
  hasCronSchedule,
  orgId,
  installId,
}: {
  runnerStatus?: string
  kind: TInstallCronOfflineKind
  hasCronSchedule?: boolean
  orgId?: string
  installId?: string
}) => {
  if (runnerStatus !== 'offline' || !hasCronSchedule) return null

  const { title, description } = COPY[kind]

  return (
    <Banner theme="warn">
      <div className="flex items-center justify-between gap-4">
        <div className="flex flex-col gap-0.5">
          <Text weight="strong">{title}</Text>
          <Text variant="subtext">{description}</Text>
        </div>
      </div>
    </Banner>
  )
}
