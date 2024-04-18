import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { Heading, Grid, OrgCard, Page, PageHeader } from '@/components'
import { getOrgs } from '@/lib'

export default withPageAuthRequired(
  async function OrgDashboard() {
    const orgs = await getOrgs()

    return (
      <Page
        header={
          <PageHeader
            title={
              <Heading level={1} variant="title">
                Your orgs
              </Heading>
            }
          />
        }
      >
        <Grid>{orgs?.map((o) => <OrgCard key={o?.id} {...o} />)}</Grid>
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
