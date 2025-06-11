import { cookies, headers } from 'next/headers'
import { notFound } from 'next/navigation'
import type { FC } from 'react'
import {
  withPageAuthRequired,
  type AppRouterPageRoute,
} from '@auth0/nextjs-auth0'
import { PageProvider, InstallProvider } from '@/stratus/context'
import type { ILayoutProps, TInstall } from '@/types'
import { nueQueryData } from '@/utils'

const StratusInstallLayout: FC<ILayoutProps<'org-id' | 'install-id'>> = async ({
  children,
  params,
}) => {
  const cookieStore = cookies()
  const isPageNavOpen = Boolean(
    cookieStore.get('is-page-nav-open')?.value === 'true'
  )
  const orgId = params?.['org-id']
  const installId = params?.['install-id']
  const { data, error } = await nueQueryData<TInstall>({
    orgId,
    path: `installs/${installId}`,
  })

  if (error) {
    notFound()
  }

  return (
    <InstallProvider initInstall={data} shouldPoll>
      <PageProvider initIsPageNavOpen={isPageNavOpen}>{children}</PageProvider>
    </InstallProvider>
  )
}

export default withPageAuthRequired(
  StratusInstallLayout as AppRouterPageRoute,
  {
    returnTo() {
      return headers().get('x-origin-path')
    },
  }
)
