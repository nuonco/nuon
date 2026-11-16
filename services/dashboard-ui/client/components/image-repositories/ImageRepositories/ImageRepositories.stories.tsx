import { PageStory } from '@/components/cloud-connections/__stories__/fixtures'
import { repositories } from '../fixtures'
import { ImageRepositories } from './ImageRepositories'

export default { title: 'Features / Image repositories / List' }

export const Populated = () => (
  <PageStory>
    <ImageRepositories orgId="org-mock-001" repositories={repositories} />
  </PageStory>
)
export const Empty = () => (
  <PageStory>
    <ImageRepositories orgId="org-mock-001" repositories={[]} />
  </PageStory>
)
export const Loading = () => (
  <PageStory>
    <ImageRepositories orgId="org-mock-001" repositories={[]} isLoading />
  </PageStory>
)
export const Failed = () => (
  <PageStory>
    <ImageRepositories
      orgId="org-mock-001"
      repositories={[]}
      error={{
        error: 'Repository loading failed',
        description: 'Try refreshing the page.',
        user_error: false,
      }}
    />
  </PageStory>
)
