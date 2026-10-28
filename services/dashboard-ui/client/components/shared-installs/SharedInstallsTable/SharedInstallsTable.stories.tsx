export default {
  title: 'Playground / Customer portal / Shared installs table',
}

import { mockSharedInstallRows } from '../fixtures'
import { SharedInstallsTable } from './SharedInstallsTable'

export const Empty = () => (
  <div className="p-4">
    <SharedInstallsTable rows={[]} />
  </div>
)

export const Populated = () => (
  <div className="p-4">
    <SharedInstallsTable rows={mockSharedInstallRows} />
  </div>
)

export const Loading = () => (
  <div className="p-4">
    <SharedInstallsTable rows={[]} isLoading />
  </div>
)

export const ErrorState = () => (
  <div className="p-4">
    <SharedInstallsTable rows={[]} isError />
  </div>
)
