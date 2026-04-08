export default {
  title: 'Installs/InstallConfigGraph',
}

import { InstallConfigGraph } from './InstallConfigGraph'

export const Default = () => (
  <InstallConfigGraph appId="app-1" appConfigId="config-1" />
)

export const MissingData = () => (
  <InstallConfigGraph appId={undefined} appConfigId={undefined} />
)
