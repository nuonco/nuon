import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import type { TApp, TAppSandboxBuild } from '@/types'

interface ISandboxBuildHeader {
  app: TApp
  build: TAppSandboxBuild
  orgId: string
}

export const SandboxBuildHeader = ({
  app,
  build,
  orgId,
}: ISandboxBuildHeader) => (
  <DetailHeader
    title="Sandbox build"
    id={build?.id}
    identity={
      <Time
        time={build?.created_at}
        format="relative"
        variant="subtext"
        theme="info"
      />
    }
    actions={
      build?.runner_job?.id ? (
        <Button href={`/${orgId}/runner/jobs/${build.runner_job.id}`} variant="secondary">
          View execution
        </Button>
      ) : null
    }
    metadata={
      <>
        <LabeledStatus
          label="Status"
          statusProps={{
            status: build?.status_v2?.status ?? build?.status,
          }}
          tooltipProps={{
            tipContentClassName: 'w-fit',
            tipContent: (
              <Text nowrap variant="subtext">
                {build?.status_v2?.status_human_description ??
                  build?.status_description}
              </Text>
            ),
            position: 'bottom',
          }}
        />
        <LabeledValue label="Duration">
          <Duration
            variant="subtext"
            beginTime={build?.created_at}
            endTime={build?.updated_at}
          />
        </LabeledValue>
        <LabeledValue label="App">
          <Link href={`/${orgId}/apps/${app?.id}`}>{app?.name}</Link>
        </LabeledValue>
        <LabeledValue label="Config">
          <Link href={`/${orgId}/apps/${app?.id}/sandbox`}>
            {app?.name} sandbox
          </Link>
        </LabeledValue>
        <LabeledValue label="Built by">
          {build?.created_by?.email ? (
            <Text variant="subtext">{build.created_by.email}</Text>
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          )}
        </LabeledValue>
      </>
    }
  />
)
