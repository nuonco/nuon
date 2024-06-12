import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Heading,
  Grid,
  Link,
  OrgCard,
  Page,
  PageHeader,
  Text,
} from '@/components'
import { OrgProvider } from '@/context'
import { getOrgs } from '@/lib'

export default withPageAuthRequired(
  async function OrgDashboard() {
    const orgs = await getOrgs()
    const hasOrgs = Boolean(orgs)

    return (
      <Page
        header={
          <PageHeader
            title={
              <Heading level={1} variant="title">
                {hasOrgs ? 'Your organizations' : 'Welcome to Nuon'}
              </Heading>
            }
          />
        }
      >
        {hasOrgs ? (
          <Grid>
            {orgs?.map((o) => (
              <OrgProvider key={o?.id} initOrg={o}>
                <OrgCard />
              </OrgProvider>
            ))}
          </Grid>
        ) : (
          <div className="max-w-lg flex flex-col gap-4">
            <Heading variant="subtitle">
              You need to create an organization
            </Heading>
            <Text className="inline-flex">
              To create your organization and get started with Nuon please
              contact us at{' '}
              <Link className="inline-flex" href="mailto:team@nuon.co">
                team@nuon.co
              </Link>
              .
            </Text>
          </div>
        )}
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
