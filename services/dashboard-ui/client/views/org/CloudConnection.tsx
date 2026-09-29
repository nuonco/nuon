import { useParams } from 'react-router'
import {
  ConnectionDetail,
  type TConnectionTab,
} from '@/components/cloud-connections/ConnectionDetail'
import { useCloudConnection } from '@/components/cloud-connections/queries'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const CloudConnection = ({
  tab = 'overview',
}: {
  tab?: TConnectionTab
}) => {
  const { connectionId } = useParams()
  const { org } = useOrg()
  const query = useCloudConnection(org?.id, connectionId!)
  return (
    <>
      <PageTitle
        title={
          tab === 'overview'
            ? query.data?.name || 'Cloud connection'
            : `${query.data?.name || 'Cloud connection'} ${tab}`
        }
      />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/settings`, text: 'Settings' },
          {
            path: `/${org?.id}/settings/cloud-connections`,
            text: 'Cloud connections',
          },
          {
            path: `/${org?.id}/settings/cloud-connections/${connectionId}`,
            text: query.data?.name || 'Cloud connection',
          },
        ]}
      />
      <ConnectionDetail
        orgId={org?.id}
        connectionId={connectionId!}
        connection={query.data}
        isVerifying={query.isVerifying}
        verificationTimedOut={query.verificationTimedOut}
        error={query.error}
        tab={tab}
      />
    </>
  )
}
