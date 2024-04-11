import { Card, Grid, Install, Link, OrgHeading, Page, Text } from '@/components'
import { getInstalls, getOrg } from '@/lib'

export default async function InstallsDashboard({ params }) {
  const orgId = params?.['org-id']
  const [installs, org] = await Promise.all([
    getInstalls({ orgId }),
    getOrg({ orgId }),
  ])

  return (
    <Page heading={<OrgHeading {...org} />} links={[{ href: orgId }]}>
      <Grid>
        {installs?.map((install) => (
          <Card key={install?.id}>
            <Install install={install} />
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
}
