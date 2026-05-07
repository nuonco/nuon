import { useState } from 'react'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IDownloadLogsModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  includeSystemLogs: boolean
  onDownload: (includeSystemLogs: boolean) => void
}

export const DownloadLogsModal = ({
  isPending,
  includeSystemLogs: defaultIncludeSystem,
  onDownload,
  ...props
}: IDownloadLogsModal) => {
  const [includeSystem, setIncludeSystem] = useState(defaultIncludeSystem)

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
        onClick: () => onDownload(includeSystem),
        variant: 'primary',
      }}
      {...props}
    >
      <Text variant="base">Download logs from this stream as a text file.</Text>
      <div className="flex">
        <CheckboxInput
          checked={includeSystem}
          onChange={() => setIncludeSystem(!includeSystem)}
          labelProps={{ labelText: 'Include system logs' }}
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
    <Button
      aria-label="Download logs"
      title="Download logs"
      onClick={onClick}
      {...props}
    >
      <Icon variant="FileArrowDownIcon" size="16" />
    </Button>
  )
}
