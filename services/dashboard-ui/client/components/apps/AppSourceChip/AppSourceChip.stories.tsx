export default {
  title: 'Features / Apps / Source chip',
}

import { AppSourceChip } from './AppSourceChip'

export const Default = () => (
  <AppSourceChip
    repo="acme/platform-configs"
    repoHref="https://github.com/acme/platform-configs"
  />
)

export const Loading = () => <AppSourceChip isLoading />
