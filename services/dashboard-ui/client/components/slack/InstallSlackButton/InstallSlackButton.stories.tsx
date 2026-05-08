import { InstallSlackButton } from './InstallSlackButton'

export default { title: 'Slack/InstallSlackButton' }

export const Default = () => (
  <InstallSlackButton isPending={false} onInstall={() => {}} />
)

export const Pending = () => (
  <InstallSlackButton isPending={true} onInstall={() => {}} />
)
