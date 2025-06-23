import { cookies } from 'next/headers'
import { notFound } from 'next/navigation'
import type { FC } from 'react'
import { Dashboard } from '@/stratus/components'
import { DashboardProvider, OrgProvider } from '@/stratus/context'
import type { ILayoutProps, TOrg } from '@/types'
import { nueQueryData } from '@/utils'

const StratusLayout: FC<ILayoutProps<'org-id'>> = async ({
  children,
  params,
}) => {
  const cookieStore = await cookies()
  const isSidebarOpen = Boolean(
    cookieStore.get('is-sidebar-open')?.value === 'true'
  )
  const { ['org-id']: orgId } = await params
  const { data: org, error } = await nueQueryData<TOrg>({
    orgId,
    path: `orgs/current`,
  })

  if (error) {
    notFound()
  }

  return (
    <OrgProvider initOrg={org} shouldPoll>
      <DashboardProvider initIsSidebarOpen={isSidebarOpen}>
        <Dashboard>{children}</Dashboard>
      </DashboardProvider>
    </OrgProvider>
  )
}

export default StratusLayout
