export default {
  title: 'LogStream/DownloadLogs',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DownloadLogsModal, DownloadLogsButton } from './DownloadLogs'

export const Default = () => (
  <ModalStory>
    <DownloadLogsModal
      isPending={false}
      onDownload={(mode) => alert(`Downloading: ${mode}`)}
    />
  </ModalStory>
)

export const Downloading = () => (
  <ModalStory>
    <DownloadLogsModal
      isPending={true}
      onDownload={() => {}}
    />
  </ModalStory>
)

export const Button = () => (
  <DownloadLogsButton onClick={() => alert('Download clicked')} />
)
