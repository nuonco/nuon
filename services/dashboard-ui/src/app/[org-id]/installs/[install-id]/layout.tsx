import { getInstallById } from '@/lib'
import { InstallProvider } from '@/providers/install-provider'
import type { TLayoutProps } from '@/types'

interface IInstallLayout extends TLayoutProps<'org-id' | 'install-id'> {}

export default async function InstallLayout({
  children,
  params,
}: IInstallLayout) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install, error } = await getInstallById({
    orgId,
    installId,
  })

  if (error) {
    console.error('error fetching install by id', error)
  }

  return (
    <InstallProvider initInstall={install} shouldPoll>
      {children}
    </InstallProvider>
  )
}
