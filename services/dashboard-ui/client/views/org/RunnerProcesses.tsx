import { ListPage } from '@/components/layout/ListPage'
import { PageLayout } from '@/components/layout/PageLayout'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunnerProcessesTable } from '@/components/runners/RunnerProcessesTable'
import { useOrg } from '@/hooks/use-org'
import { RunnerProvider } from '@/providers/runner-provider'

export const RunnerProcesses = () => {
  const { org } = useOrg()
  const runnerId = org?.runner_group?.runners?.[0]?.id

  const breadcrumbs = [
    { path: `/${org.id}`, text: org?.name },
    { path: `/${org.id}/runner`, text: 'Build runner' },
    { path: `/${org.id}/runner/processes`, text: 'Processes' },
  ]

  if (!runnerId) {
    return (
      <>
        <PageTitle title="Runner processes" />
        <Breadcrumbs breadcrumbs={breadcrumbs} />
        <PageLayout>
          <SectionHeader
            variant="page"
            title="Runner processes"
            description="No build runner configured."
          />
        </PageLayout>
      </>
    )
  }

  return (
    <RunnerProvider runnerId={runnerId} shouldPoll>
      <PageTitle title="Runner processes" />
      <Breadcrumbs breadcrumbs={breadcrumbs} />
      <ListPage
        variant="page"
        title="Runner processes"
        description="View and manage runner process lifecycle, uptime, and shutdowns."
      >
        <RunnerProcessesTable shouldPoll />
      </ListPage>
    </RunnerProvider>
  )
}
