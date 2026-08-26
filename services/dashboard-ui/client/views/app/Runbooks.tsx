import { RunbooksTable } from '@/components/runbooks/RunbooksTable'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Runbooks = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const hasRunbookStudio = !!org?.features?.['runbook-studio']

  return (
    <>
      <PageTitle segments={['Runbooks', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/runbooks`, text: 'Runbooks' },
        ]}
      />
      <ListPage
        title="App runbooks"
        description="Define and manage operational procedures for your installs."
        actions={
          hasRunbookStudio ? (
            <Button variant="primary" href={`/${org?.id}/apps/${app?.id}/studio`}>
              <Icon variant="ListChecksIcon" size={16} />
              Open runbook studio
            </Button>
          ) : null
        }
      >
        <RunbooksTable />
      </ListPage>
    </>
  )
}
