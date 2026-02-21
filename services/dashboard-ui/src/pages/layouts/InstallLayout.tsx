import { Outlet, useParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'
import { InstallContext } from '@/providers/install-provider'
import type { TInstall } from '@/types'

export default function InstallLayout() {
  const { installId } = useParams()
  const { org } = useOrg()

  const {
    data: install,
    error,
    isLoading,
  } = usePolling<TInstall>({
    path: `/api/orgs/${org?.id}/installs/${installId}`,
    shouldPoll: !!org?.id && !!installId,
    pollInterval: 20000,
  })

  if (!install && isLoading) {
    return (
      <div className="flex items-center justify-center h-full p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900" />
      </div>
    )
  }

  return (
    <InstallContext.Provider
      value={{
        install: install || null,
        isLoading,
        error,
        refresh: () => {},
      }}
    >
      <Outlet />
    </InstallContext.Provider>
  )
}
