import { useMutation } from '@tanstack/react-query'
import type { IButtonAsButton } from '@/components/common/Button'
import type { IModal } from '@/components/surfaces/Modal'
import { useLogStreamData } from '@/hooks/use-logs'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createFileDownload } from '@/utils/file-download'
import { DownloadLogsModal, DownloadLogsButton } from './DownloadLogs'

interface IDownloadLogsModalContainer extends IModal {
  orgId: string
  logStreamId: string
  runnerJobId?: string
  includeSystemLogs: boolean
}

export const DownloadLogsModalContainer = ({
  orgId,
  logStreamId,
  runnerJobId,
  includeSystemLogs,
  ...props
}: IDownloadLogsModalContainer) => {
  const { removeModal } = useSurfaces()

  const { mutate: download, isPending } = useMutation({
    mutationFn: async (includeSystem: boolean) => {
      const params = new URLSearchParams()
      if (!includeSystem) params.set('job_output', 'true')
      if (runnerJobId) params.set('runner_job_id', runnerJobId)
      const query = params.toString()
      const resp = await fetch(
        `/api/orgs/${orgId}/log-streams/${logStreamId}/logs/download${query ? `?${query}` : ''}`
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
      includeSystemLogs={includeSystemLogs}
      onDownload={(includeSystem) => download(includeSystem)}
      {...props}
    />
  )
}

export const DownloadLogsButtonContainer = ({
  includeSystemLogs = false,
  onClick: _onClick,
  ...props
}: { includeSystemLogs?: boolean } & IButtonAsButton) => {
  const { org } = useOrg()
  const { logStreamId, runnerJobId } = useLogStreamData()
  const { addModal } = useSurfaces()
  const modal = (
    <DownloadLogsModalContainer
      orgId={org.id}
      logStreamId={logStreamId}
      runnerJobId={runnerJobId}
      includeSystemLogs={includeSystemLogs}
    />
  )

  return <DownloadLogsButton onClick={() => addModal(modal)} {...props} />
}
