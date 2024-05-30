import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { Heading, Grid, OrgCard, Page, PageHeader } from '@/components'
import { OrgProvider } from '@/context'
import { getOrgs } from '@/lib'
import { createOrg } from './actions'

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
                {hasOrgs ? 'Your organizations' : 'Create your organization'}
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
          <form className="flex flex-col gap-4 max-w-md" action={createOrg}>
            <label className="flex flex-col flex-auto gap-2">
              <span className="font-semibold">Organization name</span>
              <input
                className="border bg-inherit rounded px-4 py-1.5 shadow-inner"
                name="name"
                type="text"
                required
              />
            </label>

            <button className="rounded text-sm text-gray-50 bg-fuchsia-600 hover:bg-fuchsia-700 focus:bg-fuchsia-700 active:bg-fuchsia-800 px-4 py-1.5 w-fit">
              Create organization
            </button>
          </form>
        )}
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
