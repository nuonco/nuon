import { PrincipalSharing } from './PrincipalSharing'
import { connection, customConnection, PageStory } from './__stories__/fixtures'

export default { title: 'Features / Cloud connections / Shared principal' }

export const Shared = () => (
  <PageStory>
    <PrincipalSharing
      connections={[connection, customConnection]}
      orgId="org-mock-001"
    />
  </PageStory>
)

export const Dedicated = () => <PrincipalSharing connections={[]} />
