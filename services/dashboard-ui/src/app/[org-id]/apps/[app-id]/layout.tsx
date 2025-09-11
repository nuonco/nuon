import { getAppById } from '@/lib'
import { AppProvider } from '@/providers/app-provider'
import type { TLayoutProps } from '@/types'

interface IAppLayout extends TLayoutProps<'org-id' | 'app-id'> {}

export default async function AppLayout({ children, params }: IAppLayout) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app, error } = await getAppById({
    orgId,
    appId,
  })

  if (error) {
    console.error('error fetching app by id', error)
  }

  return (
    <AppProvider initApp={app} shouldPoll>
      {children}
    </AppProvider>
  )
}
