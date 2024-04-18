import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Card,
  Grid,
  Install,
  Link,
  OrgPageHeader,
  Page,
  Text,
} from '@/components'
import { getInstalls, getOrg } from '@/lib'

export default withPageAuthRequired(
  async function InstallsDashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const [installs, org] = await Promise.all([
      getInstalls({ orgId }),
      getOrg({ orgId }),
    ])

    return (
      <Page header={<OrgPageHeader {...org} />} links={[{ href: orgId }]}>
        <Grid>
          {installs?.map((install) => (
            <Card key={install?.id}>
              <Install install={install} orgId={orgId} />
              <Text variant="caption">
                <Link href={`/dashboard/${install?.org_id}/${install?.id}`}>
                  Details
                </Link>
              </Text>
            </Card>
          ))}
        </Grid>
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
