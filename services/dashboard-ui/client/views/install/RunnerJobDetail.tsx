import { useParams } from 'react-router'
import { RunnerJobHeader } from '@/components/runners/job-details/RunnerJobHeader'
import { RunnerJobLogs } from '@/components/runners/job-details/RunnerJobLogs'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunnerJobProvider } from '@/providers/runner-job-provider'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useRunnerJob } from '@/hooks/use-runner-job'
import { getJobName } from '@/utils/runner-utils'

const RunnerJobDetailContent = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { job } = useRunnerJob()

  return (
    <PageSection isScrollable className="!p-0 !gap-0">
      <PageTitle title={`Job | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          { path: `/${org?.id}/installs/${install?.id}/runner`, text: 'Runner' },
          { path: '', text: getJobName(job) },
        ]}
      />
      <RunnerJobHeader />
      <PageSection isScrollable={false} className="!pb-12">
        <RunnerJobLogs />
      </PageSection>
    </PageSection>
  )
}

export const RunnerJobDetail = () => {
  const { jobId } = useParams()

  return (
    <RunnerJobProvider runnerJobId={jobId!}>
      <RunnerJobDetailContent />
    </RunnerJobProvider>
  )
}
