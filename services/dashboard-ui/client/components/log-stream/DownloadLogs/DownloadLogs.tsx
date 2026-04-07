import { useState } from 'react'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'

type DownloadMode = 'all' | 'user'

interface IDownloadLogsModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  onDownload: (mode: DownloadMode) => void
}

export const DownloadLogsModal = ({
  isPending,
  onDownload,
  ...props
}: IDownloadLogsModal) => {
  const [mode, setMode] = useState<DownloadMode>('all')

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
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
        onClick: () => onDownload(mode),
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

export const DownloadLogsButton = ({
  onClick,
  ...props
}: { onClick: () => void } & Omit<IButtonAsButton, 'onClick'>) => {
  return (
    <Button variant="ghost" onClick={onClick} {...props}>
      Download
      <Icon variant="FileArrowDownIcon" size="16" />
    </Button>
  )
}
