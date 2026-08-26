import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunSummary } from '@/components/runs/RunSummary'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSandboxRun } from '@/hooks/use-sandbox-run'

export const SandboxRunSummaryTab = () => {
  const { sandboxRun } = useSandboxRun()
  const { install } = useInstall()
  const { org } = useOrg()

  return (
    <>
      <PageTitle segments={['Sandbox run', install?.name]} />
      <RunSummary
        isLoading={!sandboxRun}
        status={sandboxRun?.status_v2}
        statusDescription={sandboxRun?.status_description}
        timings={[
          { label: 'Created', time: sandboxRun?.created_at },
          { label: 'Planned', time: sandboxRun?.planned_at },
          { label: 'Applied', time: sandboxRun?.applied_at },
        ]}
        duration={{
          beginTime: sandboxRun?.created_at,
          endTime: sandboxRun?.updated_at,
        }}
        jobs={sandboxRun?.runner_jobs}
        jobHref={(job) =>
          `/${org?.id}/installs/${install?.id}/runner/jobs/${job?.id}`
        }
        triggeredBy={
          sandboxRun?.created_by?.email ? (
            <Text variant="subtext">{sandboxRun.created_by.email}</Text>
          ) : sandboxRun?.created_by_id ? (
            <ID>{sandboxRun.created_by_id}</ID>
          ) : null
        }
      />
    </>
  )
}
