import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useLogStream } from '@/hooks/use-log-stream'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createFileDownload } from '@/utils/file-download'

type DownloadMode = 'all' | 'user'

interface IDownloadLogsModal extends IModal {
  orgId: string
  logStreamId: string
}

export const DownloadLogsModal = ({
  orgId,
  logStreamId,
  ...props
}: IDownloadLogsModal) => {
  const { removeModal } = useSurfaces()
  const [mode, setMode] = useState<DownloadMode>('all')

  const { mutate: download, isPending } = useMutation({
    mutationFn: async () => {
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
    <Modal
      heading={
        <Text
          className="!flex items-center gap-4"
          variant="h3"
          theme="info"
          weight="strong"
        >
          <Icon variant="FileArrowDownIcon" size={24} /> Download logs
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Downloading...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="FileArrowDownIcon" size="18" /> Download logs
          </span>
        ),
        disabled: isPending,
        onClick: () => download(),
        variant: 'primary',
      }}
      {...props}
    >
      <Text variant="base">Download logs from this stream as a text file.</Text>
      <div className="flex items-center gap-4">
        <RadioInput
          name="download-mode"
          value="all"
          checked={mode === 'all'}
          onChange={() => setMode('all')}
          labelProps={{ labelText: 'All logs' }}
        />
        <RadioInput
          name="download-mode"
          value="user"
          checked={mode === 'user'}
          onChange={() => setMode('user')}
          labelProps={{ labelText: 'Job output only' }}
        />
      </div>
    </Modal>
  )
}

export const DownloadLogsButton = ({ ...props }: IButtonAsButton) => {
  const { org } = useOrg()
  const { logStream } = useLogStream()
  const { addModal } = useSurfaces()
  const modal = <DownloadLogsModal orgId={org.id} logStreamId={logStream.id} />

  return (
    <Button variant="ghost" onClick={() => addModal(modal)} {...props}>
      Download
      <Icon variant="FileArrowDownIcon" size="16" />
    </Button>
  )
}
