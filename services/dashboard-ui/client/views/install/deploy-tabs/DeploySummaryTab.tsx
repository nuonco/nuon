import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunSummary } from '@/components/runs/RunSummary'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const DeploySummaryTab = () => {
  const { deploy } = useDeploy()
  const { install } = useInstall()
  const { org } = useOrg()

  return (
    <>
      <PageTitle segments={['Deploy', install?.name]} />
      <RunSummary
        isLoading={!deploy}
        status={deploy?.status_v2}
        statusDescription={deploy?.status_description}
        timings={[
          { label: 'Created', time: deploy?.created_at },
          { label: 'Planned', time: deploy?.planned_at },
          { label: 'Applied', time: deploy?.applied_at },
        ]}
        duration={{
          beginTime: deploy?.created_at,
          endTime: deploy?.updated_at,
        }}
        jobs={deploy?.runner_jobs}
        jobHref={(job) =>
          `/${org?.id}/installs/${install?.id}/runner/jobs/${job?.id}`
        }
        triggeredBy={
          deploy?.created_by?.email ? (
            <Text variant="subtext">{deploy.created_by.email}</Text>
          ) : deploy?.created_by_id ? (
            <ID>{deploy.created_by_id}</ID>
          ) : null
        }
      />
    </>
  )
}
