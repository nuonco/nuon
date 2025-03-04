// @ts-nocheck
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { headers } from 'next/headers'
import { notFound } from 'next/navigation'
import { Layout } from '@/components'
import { getAPIVersion, getOrg, getOrgs } from '@/lib'
import {
  ORG_DASHBOARD,
  ORG_RUNNER,
  ORG_SETTINGS,
  ORG_SUPPORT,
  VERSION,
} from '@/utils'

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
          ORG_DASHBOARD,
          ORG_RUNNER,
          ORG_SETTINGS,
          ORG_SUPPORT,
        }}
      >
        {children}
      </Layout>
    )
  },
  {
    returnTo() {
      return headers().get('x-origin-path')
    },
  }
)
