export default {
  title: 'Layout/PageContent',
}

import { PageContent } from './PageContent'
import { Text } from '@/components/common/Text'

export const Column = () => (
  <PageContent variant="column">
    <div className="p-4 border rounded">
      <Text>Column item 1</Text>
    </div>
    <div className="p-4 border rounded">
      <Text>Column item 2</Text>
    </div>
  </PageContent>
)

export const Row = () => (
  <PageContent variant="row">
    <div className="p-4 border rounded">
      <Text>Row item 1</Text>
    </div>
    <div className="p-4 border rounded">
      <Text>Row item 2</Text>
    </div>
  </PageContent>
)
