import { DownloadBundle } from './DownloadBundle'

export default {
  title: 'Apps/Bundles/DownloadBundle',
}

export const Default = () => (
  <DownloadBundle isPending={false} onClick={() => {}} />
)

export const Pending = () => (
  <DownloadBundle isPending={true} onClick={() => {}} />
)
