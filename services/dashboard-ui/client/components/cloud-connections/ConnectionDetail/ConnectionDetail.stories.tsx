import { ConnectionDetail } from './ConnectionDetail'
import { InstallsTableComponent } from '@/components/installs/InstallsTable'
import {
  PageStory,
  connection,
  customConnection,
  failedConnection,
  verifiedConnection,
  noop,
} from '../__stories__/fixtures'
export default { title: 'Features / Cloud connections / Detail' }
const props = {
  basePath: '/org-mock-001/cloud-connections/cc-example',
  onVerify: noop,
  onDelete: noop,
}
export const Overview = () => (
  <PageStory>
    <ConnectionDetail {...props} connection={connection} tab="overview" />
  </PageStory>
)
export const Custom = () => (
  <PageStory>
    <ConnectionDetail {...props} connection={customConnection} tab="overview" />
  </PageStory>
)
export const Loading = () => (
  <PageStory>
    <ConnectionDetail {...props} tab="overview" />
  </PageStory>
)
export const InstallsEmpty = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={connection}
      tab="installs"
      installs={
        <InstallsTableComponent
          data={[]}
          isLoading={false}
          pagination={{ offset: 0, limit: 20 }}
          emptyTitle="No installs use this connection yet"
          emptyMessage="Select this connection when creating an install."
        />
      }
    />
  </PageStory>
)
export const Verification = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={verifiedConnection}
      tab="verification"
    />
  </PageStory>
)
export const Failed = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={failedConnection}
      tab="verification"
    />
  </PageStory>
)
export const Verifying = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={verifiedConnection}
      tab="verification"
      isVerifying
    />
  </PageStory>
)
