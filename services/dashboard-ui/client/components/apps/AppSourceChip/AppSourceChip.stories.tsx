export default {
  title: 'Apps/AppSourceChip',
}

import { AppSourceChip } from './AppSourceChip'

export const Default = () => (
  <AppSourceChip
    repo="acme/platform-configs"
    repoHref="https://github.com/acme/platform-configs"
    connectHref="/org-1/apps/app-1"
  />
)

export const NoRepo = () => <AppSourceChip connectHref="/org-1/apps/app-1" />

export const Loading = () => (
  <AppSourceChip isLoading connectHref="/org-1/apps/app-1" />
)
