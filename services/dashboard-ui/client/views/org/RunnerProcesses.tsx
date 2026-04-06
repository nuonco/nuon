import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunnerProcessesTable } from '@/components/runners/RunnerProcessesTable'
import { useOrg } from '@/hooks/use-org'
import { RunnerProvider } from '@/providers/runner-provider'

const CONTAINER_ID = 'org-runner-processes-page'

export const RunnerProcesses = () => {
  const { org } = useOrg()
  const runnerId = org?.runner_group?.runners?.[0]?.id

  if (!runnerId) {
    return (
      <PageLayout>
        <PageTitle title={`Runner processes | ${org?.name}`} />
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${org.id}`, text: org?.name },
            { path: `/${org.id}/runner`, text: 'Build runner' },
            { path: `/${org.id}/runner/processes`, text: 'Processes' },
          ]}
        />
        <PageHeader>
          <HeadingGroup>
            <Text variant="h3" weight="strong" level={1}>
              Runner processes
            </Text>
            <Text theme="neutral">No build runner configured.</Text>
          </HeadingGroup>
        </PageHeader>
      </PageLayout>
    )
  }

  return (
    <RunnerProvider runnerId={runnerId} shouldPoll>
      <PageLayout className="pb-6" id={CONTAINER_ID} isScrollable>
        <PageTitle title={`Runner processes | ${org?.name}`} />
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${org.id}`, text: org?.name },
            { path: `/${org.id}/runner`, text: 'Build runner' },
            { path: `/${org.id}/runner/processes`, text: 'Processes' },
          ]}
        />
        <PageHeader>
          <HeadingGroup>
            <Text variant="h3" weight="strong" level={1}>
              Runner processes
            </Text>
            <Text theme="neutral">
              View and manage runner process lifecycle, uptime, and shutdowns.
            </Text>
          </HeadingGroup>
        </PageHeader>
        <PageContent>
          <PageSection>
            <RunnerProcessesTable shouldPoll />
          </PageSection>
        </PageContent>
        <BackToTop containerId={CONTAINER_ID} />
      </PageLayout>
    </RunnerProvider>
  )
}
