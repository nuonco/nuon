import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallState } from '@/lib'
import type { TAPIError } from '@/types'
import { createFileDownload } from '@/utils/file-download'
import { InstallState } from './InstallState'

export const InstallStateContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const {
    data: state,
    error,
    isLoading,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-state', org?.id, install?.id],
    queryFn: () => getInstallState({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const value = state ? JSON.stringify(state, null, 2) : undefined
  const filename = `${install?.name || 'install'}-state.json`

  return (
    <InstallState
      value={value}
      filename={filename}
      loading={isLoading}
      error={
        error
          ? (error as TAPIError).error || 'Unable to load install state.'
          : undefined
      }
      onDownload={
        value
          ? () => createFileDownload(value, filename, 'application/json')
          : undefined
      }
    />
  )
}
