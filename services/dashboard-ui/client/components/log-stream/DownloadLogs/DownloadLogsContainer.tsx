import { useMutation } from '@tanstack/react-query'
import type { IButtonAsButton } from '@/components/common/Button'
import type { IModal } from '@/components/surfaces/Modal'
import { useLogStream } from '@/hooks/use-log-stream'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createFileDownload } from '@/utils/file-download'
import { DownloadLogsModal, DownloadLogsButton } from './DownloadLogs'

interface IDownloadLogsModalContainer extends IModal {
  orgId: string
  logStreamId: string
}

export const DownloadLogsModalContainer = ({
  orgId,
  logStreamId,
  ...props
}: IDownloadLogsModalContainer) => {
  const { removeModal } = useSurfaces()

  const { mutate: download, isPending } = useMutation({
    mutationFn: async (mode: 'all' | 'user') => {
      const params = mode === 'user' ? '?job_output=true' : ''
      const resp = await fetch(
        `/api/orgs/${orgId}/log-streams/${logStreamId}/logs/download${params}`
      )
      if (!resp.ok) {
        throw new Error('Failed to download logs')
      }
      return resp.blob()
    },
    onSuccess: (blob) => {
      createFileDownload(blob, `logs-${logStreamId}.txt`, 'text/plain')
      removeModal(props.modalId)
    },
  })

  return (
    <DownloadLogsModal
      isPending={isPending}
      onDownload={(mode) => download(mode)}
      {...props}
    />
  )
}

export const DownloadLogsButtonContainer = ({ onClick: _onClick, ...props }: IButtonAsButton) => {
  const { org } = useOrg()
  const { logStream } = useLogStream()
  const { addModal } = useSurfaces()
  const modal = <DownloadLogsModalContainer orgId={org.id} logStreamId={logStream.id} />

  return (
    <DownloadLogsButton onClick={() => addModal(modal)} {...props} />
  )
}
