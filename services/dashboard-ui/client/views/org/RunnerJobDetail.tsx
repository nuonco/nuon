import { useParams } from 'react-router'
import { BackToTop } from '@/components/common/BackToTop'
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

const CONTAINER_ID = 'org-runner-job-page'

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
      <PageSection id={CONTAINER_ID} isScrollable className="!p-0 !gap-0">       
        <RunnerJobHeader />
        <PageSection isScrollable={false} className="!pb-12">
          <RunnerJobLogs />
          <BackToTop containerId={CONTAINER_ID} />
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
