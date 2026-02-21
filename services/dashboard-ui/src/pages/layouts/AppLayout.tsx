import { Outlet, useParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'
import { AppContext } from '@/providers/app-provider'
import type { TApp } from '@/types'

export default function AppLayout() {
  const { appId } = useParams()
  const { org } = useOrg()

  const {
    data: app,
    error,
    isLoading,
  } = usePolling<TApp>({
    path: `/api/orgs/${org?.id}/apps/${appId}`,
    shouldPoll: !!org?.id && !!appId,
    pollInterval: 20000,
  })

  if (!app && isLoading) {
    return (
      <div className="flex items-center justify-center h-full p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900" />
      </div>
    )
  }

  return (
    <AppContext.Provider
      value={{
        app: app || null,
        isLoading,
        error,
        refresh: () => {},
      }}
    >
      <Outlet />
    </AppContext.Provider>
  )
}
