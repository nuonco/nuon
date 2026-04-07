export default {
  title: 'Layout/PageGrid',
}

import { PageGrid } from './PageGrid'
import { Text } from '@/components/common/Text'

export const Default = () => (
  <PageGrid>
    <div className="p-6 border rounded bg-white dark:bg-dark-grey-800">
      <Text variant="h3" weight="strong">Main content</Text>
      <Text variant="body" theme="neutral">Primary content area that spans most of the width.</Text>
    </div>
    <div className="p-6 border rounded bg-white dark:bg-dark-grey-800">
      <Text variant="h3" weight="strong">Sidebar</Text>
      <Text variant="body" theme="neutral">Secondary content in the sidebar column.</Text>
    </div>
  </PageGrid>
)
