import { useParams } from 'react-router'
import { ConnectionWizard } from '@/components/cloud-connections/ConnectionWizard'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useOrg } from '@/hooks/use-org'

export const CloudConnectionCreate = () => {
  const { org } = useOrg()
  const { connectionId } = useParams()
  return (
    <>
      <PageTitle
        title={
          connectionId ? 'Set up cloud connection' : 'Create cloud connection'
        }
      />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/cloud-connections`, text: 'Cloud connections' },
          { path: '', text: connectionId ? 'Setup' : 'Create' },
        ]}
      />
      <ConnectionWizard
        key={connectionId || 'create'}
        orgId={org?.id}
        connectionId={connectionId}
      />
    </>
  )
}
