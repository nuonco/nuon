// @ts-nocheck
import { notFound } from 'next/navigation'
import { Layout, OrgProvider } from '@/components'
import { getAPIVersion, getOrg, getOrgs } from '@/lib'
import { VERSION } from '@/utils'

export default async function OrgLayout({ children, params }) {
  const { ['org-id']: orgId } = await params
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
        orgs={orgs}
        versions={{
          api: apiVersion,
          ui: {
            version: VERSION,
          },
        }}
      >
        {children}
      </Layout>
    </OrgProvider>
  )
}
