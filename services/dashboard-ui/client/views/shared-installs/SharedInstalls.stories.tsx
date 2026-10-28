export default {
  title: 'Playground / Customer portal / Shared installs list',
  meta: { fullBleed: true },
}

import { ListPage } from '@/components/layout/ListPage'
import { mockSharedInstallRows } from '@/components/shared-installs/fixtures'
import { SharedInstallsTable } from '@/components/shared-installs/SharedInstallsTable/SharedInstallsTable'
import { SharedInstallsStoryChrome } from '@/components/shared-installs/story-chrome'

const Page = ({
  rows = mockSharedInstallRows,
  isLoading = false,
}: {
  rows?: typeof mockSharedInstallRows
  isLoading?: boolean
}) => (
  <SharedInstallsStoryChrome>
    <ListPage
      variant="page"
      title="Shared installs"
      description="Installs shared with customers, and installs shared with you."
    >
      <SharedInstallsTable rows={rows} isLoading={isLoading} />
    </ListPage>
  </SharedInstallsStoryChrome>
)

export const Empty = () => <Page rows={[]} />
export const WithShares = () => <Page />
export const Loading = () => <Page rows={[]} isLoading />
