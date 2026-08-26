import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { ApiTokensTable } from '@/components/api-tokens/ApiTokensTable'
import { CreateApiTokenButton } from '@/components/api-tokens/CreateApiToken'

import { useOrg } from '@/hooks/use-org'

export const ApiTokens = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title="API tokens" />
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${org.id}`,
            text: org?.name,
          },
          {
            path: `/${org.id}/settings`,
            text: 'Settings',
          },
          {
            path: `/${org.id}/settings/api-tokens`,
            text: 'API tokens',
          },
        ]}
      />
      <ListPage
        title="API tokens"
        description="Manage static tokens for accessing the Nuon API in this org."
        createAction={<CreateApiTokenButton variant="primary" />}
      >
        <ApiTokensTable shouldPoll />
      </ListPage>
    </>
  )
}
