// @ts-nocheck
import { cookies } from 'next/headers'
import { notFound } from 'next/navigation'
import { Layout, OrgProvider } from '@/components'
import { REFRESH_PAGE_INTERVAL, REFRESH_PAGE_WARNING } from '@/configs/app'
import { getAPIVersion, getOrg, getOrgs } from '@/lib'
import { AutoRefreshProvider } from '@/providers/auto-refresh-provider'
import { VERSION } from '@/utils'

export default async function OrgLayout({ children, params }) {
  const cookieStore = await cookies()
  const isSidebarOpen = Boolean(
    cookieStore.get('is-sidebar-open')?.value === 'true'
  )
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
    <AutoRefreshProvider
      refreshIntervalMs={REFRESH_PAGE_INTERVAL}
      showWarning={REFRESH_PAGE_WARNING}
      warningTimeMs={30 * 1000} // 30 second warning
    >
      <OrgProvider initOrg={org} shouldPoll>
        <Layout
          isSidebarOpen={isSidebarOpen}
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
    </AutoRefreshProvider>
  )
}
