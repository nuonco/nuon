import { cookies } from 'next/headers'
import { notFound } from 'next/navigation'
import type { ReactNode } from 'react'
import { PageProvider, InstallProvider } from '@/stratus/context'
import type { TInstall } from '@/types'
import { nueQueryData } from '@/utils'

const InstallLayout = async ({
  children,
  params,
}: {
  children: ReactNode
  params: Promise<{ 'org-id': string; 'install-id': string }>
}) => {
  const cookieStore = await cookies()
  const isPageNavOpen = Boolean(
    cookieStore.get('is-page-nav-open')?.value === 'true'
  )
  const { ['install-id']: installId, ['org-id']: orgId } = await params
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

export default InstallLayout
