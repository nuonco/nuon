import { cookies, headers } from 'next/headers'
import { notFound } from 'next/navigation'
import type { FC } from 'react'
import {
  withPageAuthRequired,
  type AppRouterPageRoute,
} from '@auth0/nextjs-auth0'
import { Dashboard } from '@/stratus/components'
import { DashboardProvider, OrgProvider } from '@/stratus/context'
import type { ILayoutProps, TOrg } from '@/types'
import { nueQueryData } from '@/utils'

const StratusLayout: FC<ILayoutProps<'org-id'>> = async ({
  children,
  params,
}) => {
  const cookieStore = cookies()
  const isSidebarOpen = Boolean(
    cookieStore.get('is-sidebar-open')?.value === 'true'
  )
  const orgId = params?.['org-id']

  const { data: org, error } = await nueQueryData<TOrg>({
    orgId,
    path: `orgs/current`,
  })

  if (error) {
    notFound()
  }

  return (
    <OrgProvider initOrg={org}>
      <DashboardProvider initIsSidebarOpen={isSidebarOpen}>
        <Dashboard>{children}</Dashboard>
      </DashboardProvider>
    </OrgProvider>
  )
}

export default withPageAuthRequired(StratusLayout as AppRouterPageRoute, {
  returnTo() {
    return headers().get('x-origin-path')
  },
})
