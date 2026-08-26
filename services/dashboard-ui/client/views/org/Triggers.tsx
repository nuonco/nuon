import { CreateTriggerButton, TriggersTable } from '@/components/triggers'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const Triggers = () => {
  const { org } = useOrg()
  return (
    <>
      <PageTitle title="Triggers" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/settings`, text: 'Settings' },
          { path: `/${org?.id}/settings/triggers`, text: 'Triggers' },
        ]}
      />
      <ListPage
        title="Triggers"
        description="Configure inbound providers and inspect their trigger activity."
        createAction={<CreateTriggerButton />}
      >
        <TriggersTable />
      </ListPage>
    </>
  )
}
