import { useParams } from 'react-router'
import { RunnerJobHeader } from '@/components/runners/job-details/RunnerJobHeader'
import { RunnerJobLogs } from '@/components/runners/job-details/RunnerJobLogs'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunnerJobProvider } from '@/providers/runner-job-provider'
import { useOrg } from '@/hooks/use-org'
import { useRunnerJob } from '@/hooks/use-runner-job'
import { getJobName } from '@/utils/runner-utils'

const RunnerJobDetailContent = () => {
  const { org } = useOrg()
  const { job } = useRunnerJob()

  return (
    <PageLayout className="pb-6">
      <PageTitle title={`Job | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/runner`, text: 'Build runner' },
          { path: '', text: getJobName(job) },
        ]}
      />
      <PageSection flush>
        <RunnerJobHeader />
        <PageSection className="!pb-12">
          <RunnerJobLogs />
        </PageSection>
      </PageSection>
    </PageLayout>
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
