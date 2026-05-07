export default {
  title: 'LogStream/DownloadLogs',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DownloadLogsModal, DownloadLogsButton } from './DownloadLogs'

export const Default = () => (
  <ModalStory>
    <DownloadLogsModal
      isPending={false}
      includeSystemLogs={false}
      onDownload={(includeSystem) => alert(`Downloading: includeSystem=${includeSystem}`)}
    />
  </ModalStory>
)

export const WithSystemLogs = () => (
  <ModalStory>
    <DownloadLogsModal
      isPending={false}
      includeSystemLogs={true}
      onDownload={(includeSystem) => alert(`Downloading: includeSystem=${includeSystem}`)}
    />
  </ModalStory>
)

export const Downloading = () => (
  <ModalStory>
    <DownloadLogsModal
      isPending={true}
      includeSystemLogs={false}
      onDownload={() => {}}
    />
  </ModalStory>
)

export const Button = () => (
  <DownloadLogsButton onClick={() => alert('Download clicked')} />
)
