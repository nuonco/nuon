export default {
  title: 'Installs/AppInstallsList',
}

import { AppInstallsList } from './AppInstallsList'
import { mockAppInstalls } from './AppInstallsList.fixtures'

export const Default = () => <AppInstallsList installs={mockAppInstalls} />

export const Loading = () => <AppInstallsList installs={[]} isLoading />

export const Empty = () => <AppInstallsList installs={[]} />

export const NoSearchResults = () => (
  <AppInstallsList installs={mockAppInstalls} initialSearch="not-found" />
)
