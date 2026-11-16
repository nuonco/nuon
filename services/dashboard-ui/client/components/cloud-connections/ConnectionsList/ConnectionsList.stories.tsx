import { useSearchParams } from 'react-router'
import { ConnectionsList } from './ConnectionsList'
import {
  PageStory,
  connection,
  customConnection,
  failedConnection,
  verifiedACRConnection,
  verifiedECRConnection,
  verifiedGARConnection,
  verifiedConnection,
} from '../__stories__/fixtures'
export default { title: 'Features / Cloud connections / List' }

const FilteredList = () => {
  const [params] = useSearchParams()
  const platform = params.get('platform')
  const presets = params.get('preset')?.split(',')
  const search = params.get('q')?.toLowerCase() || ''
  const connections = [
    verifiedConnection,
    { ...verifiedECRConnection, id: 'cc-ecr', name: 'acme-ecr-images' },
    verifiedGARConnection,
    verifiedACRConnection,
    { ...connection, id: 'cc-pending', name: 'acme-staging' },
    { ...customConnection, id: 'cc-custom' },
    { ...failedConnection, id: 'cc-error', name: 'acme-recovery' },
  ].filter(
    (connection) =>
      (!platform || connection.platform === platform) &&
      (!presets || presets.includes(connection.preset)) &&
      [connection.name, connection.target_id, connection.principal].some(
        (value) => value.toLowerCase().includes(search)
      )
  )
  return <ConnectionsList orgId={connection.org_id} connections={connections} />
}

export const Populated = () => (
  <PageStory>
    <FilteredList />
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
export const GAR = () => (
  <PageStory>
    <ConnectionsList
      orgId={connection.org_id}
      connections={[verifiedGARConnection]}
    />
  </PageStory>
)
export const ACR = () => (
  <PageStory>
    <ConnectionsList
      orgId={connection.org_id}
      connections={[verifiedACRConnection]}
    />
  </PageStory>
)
export const MixedClouds = () => (
  <PageStory>
    <FilteredList />
  </PageStory>
)

export const ECR = () => (
  <PageStory>
    <ConnectionsList
      orgId={connection.org_id}
      connections={[verifiedECRConnection]}
    />
  </PageStory>
)
