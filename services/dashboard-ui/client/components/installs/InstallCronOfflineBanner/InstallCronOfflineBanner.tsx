import { Banner } from '@/components/common/Banner'
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
      'The runner is offline, so scheduled action runs are not firing. They will resume automatically when the runner is back online.',
  },
  sandbox_drift: {
    title: 'Sandbox drift checks are paused',
    description:
      'The runner is offline, so scheduled sandbox drift scans are not firing. They will resume automatically when the runner is back online.',
  },
  component_drift: {
    title: 'Component drift checks are paused',
    description:
      'The runner is offline, so scheduled component drift scans are not firing. They will resume automatically when the runner is back online.',
  },
}

export const InstallCronOfflineBanner = ({
  runnerStatus,
  kind,
  hasCronSchedule,
}: {
  runnerStatus?: string
  kind: TInstallCronOfflineKind
  hasCronSchedule?: boolean
}) => {
  if (runnerStatus !== 'offline' || !hasCronSchedule) return null

  const { title, description } = COPY[kind]

  return (
    <Banner theme="warn">
      <div className="flex flex-col gap-0.5">
        <Text weight="strong">{title}</Text>
        <Text variant="subtext">{description}</Text>
      </div>
    </Banner>
  )
}
