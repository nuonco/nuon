import { InstallSlackButton } from './InstallSlackButton'

export default { title: 'Features / Slack / Install slack button' }

export const Default = () => (
  <InstallSlackButton isPending={false} onInstall={() => {}} />
)

export const Pending = () => (
  <InstallSlackButton isPending={true} onInstall={() => {}} />
)
