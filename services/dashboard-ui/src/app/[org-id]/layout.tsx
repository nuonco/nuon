// @ts-nocheck
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { headers } from 'next/headers'
import { notFound } from 'next/navigation'
import { Layout, OrgProvider } from '@/components'
import { getAPIVersion, getOrg, getOrgs } from '@/lib'
import { VERSION } from '@/utils'

export default withPageAuthRequired(
  async function OrgLayout({ children, params }) {
    const orgId = params?.['org-id'] as string
    const [org, orgs, apiVersion] = await Promise.all([
      getOrg({ orgId }).catch((error) => {
        console.error(error)
        notFound()
      }),
      getOrgs().catch((error) => {
        console.error(error)
        notFound()
      }),
      getAPIVersion().catch((error) => {
        console.error(error)
        return {
          git_ref: 'unknown',
          version: 'unknown',
        }
      }),
    ])

    return (
      <OrgProvider initOrg={org} shouldPoll>
        <Layout
          org={org}
          orgs={orgs}
          versions={{
            api: apiVersion,
            ui: {
              version: VERSION,
            },
          }}
          featureFlags={{
            ORG_DASHBOARD: org?.features?.['org-dashboard'],
            ORG_RUNNER: org?.features?.['org-runner'],
            ORG_SETTINGS: org.features?.['org-settings'],
            ORG_SUPPORT: org.features?.['org-support'],
          }}
        >
          {children}
        </Layout>
      </OrgProvider>
    )
  },
  {
    returnTo() {
      return headers().get('x-origin-path')
    },
  }
)
