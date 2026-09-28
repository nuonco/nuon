import { ConnectionsList } from './ConnectionsList'
import {
  PageStory,
  connection,
  failedConnection,
  verifiedConnection,
} from '../__stories__/fixtures'
export default { title: 'Cloud connections/List' }
export const Populated = () => (
  <PageStory>
    <ConnectionsList
      orgId={connection.org_id}
      connections={[
        verifiedConnection,
        { ...connection, id: 'cc-pending', name: 'acme-staging' },
        { ...failedConnection, id: 'cc-error', name: 'acme-recovery' },
      ]}
    />
  </PageStory>
)
export const Empty = () => (
  <PageStory>
    <ConnectionsList orgId={connection.org_id} connections={[]} />
  </PageStory>
)
export const Loading = () => (
  <PageStory>
    <ConnectionsList orgId={connection.org_id} connections={[]} isLoading />
  </PageStory>
)
