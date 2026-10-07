export default { title: 'Features / Installs / Health' }

import { InstallHealthPreview } from './InstallHealthPreview'

export const ResourceCanvas = () => <InstallHealthPreview state="degraded" />
export const ReadyResources = () => <InstallHealthPreview state="healthy" />
export const ProgressingResources = () => (
  <InstallHealthPreview state="progressing" />
)
export const PartialReadiness = () => <InstallHealthPreview state="degraded" />
export const ContainerFailures = () => (
  <InstallHealthPreview state="unhealthy" />
)
export const ClusterAccessError = () => (
  <InstallHealthPreview state="access-error" />
)
export const NoObservations = () => (
  <InstallHealthPreview state="no-observations" />
)
export const EmptyResources = () => <InstallHealthPreview state="empty" />
export const Loading = () => <InstallHealthPreview state="loading" />
export const StaleSnapshot = () => <InstallHealthPreview state="stale" />
export const StaleChecks = () => <InstallHealthPreview state="stale-checks" />
