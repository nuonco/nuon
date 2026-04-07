export default {
  title: 'Layout/PageSection',
}

import { PageSection } from './PageSection'
import { Text } from '@/components/common/Text'

export const Default = () => (
  <PageSection>
    <Text variant="h3" weight="strong">Section heading</Text>
    <Text variant="body" theme="neutral">Section content goes here.</Text>
  </PageSection>
)

export const Flush = () => (
  <PageSection flush>
    <div className="bg-cool-grey-100 dark:bg-dark-grey-700 p-6">
      <Text>Flush section — no padding or gap</Text>
    </div>
  </PageSection>
)
