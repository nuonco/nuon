import { headers } from 'next/headers'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Grid,
  InstallCard,
  OrgStatus,
  OrgVCSConnections,
  OrgConnectGithubLink,
  Page,
  PageHeader,
  PageSummary,
  PageTitle,
  Text,
} from '@/components'
import { InstallProvider, OrgProvider } from '@/context'
import { getInstalls, getOrg } from '@/lib'

export default withPageAuthRequired(
  async function InstallsDashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const [installs, org] = await Promise.all([
      getInstalls({ orgId }),
      getOrg({ orgId }),
    ])

    const host = headers().get('x-forwarded-host')
    const protocol = headers().get('x-forwarded-proto')
    console.log('headers', protocol + host)

    // TODO(nnnnat): REMOVE THIS
    const TEMP = new URL('/dashboard', `${protocol}://${host}`) //JSON.stringify(process.env, null, 2) //new URL('/dashboard')

    return (
      <OrgProvider initOrg={org} shouldPoll>
        <Page
          header={
            <PageHeader
              info={<OrgStatus />}
              title={<PageTitle overline={org.id} title={org.name} />}
              summary={
                <PageSummary>
                  <OrgVCSConnections />
                  <OrgConnectGithubLink />
                </PageSummary>
              }
            />
          }
        >
          <pre>{TEMP.toJSON()}</pre>
          <Grid>
            {installs?.length ? (
              installs?.map((install) => (
                <InstallProvider key={install?.id} initInstall={install}>
                  <InstallCard />
                </InstallProvider>
              ))
            ) : (
              <Text variant="label">No installs to show</Text>
            )}
          </Grid>
        </Page>
      </OrgProvider>
    )
  },
  { returnTo: '/dashboard' }
)
