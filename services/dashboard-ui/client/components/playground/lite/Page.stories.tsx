import { Page } from './Page'
import { PlaceholderGrid } from './PlaceholderGrid'
import { appTabs } from './nav'

export default {
  title: 'Playground/Lite/Page',
}

export const ListPage = () => (
  <div className="flex flex-col gap-8 p-4">
    <Page crumbs={[{ label: 'Installs' }]} actions={['Create install']}>
      <PlaceholderGrid rows={8} />
    </Page>
  </div>
)

export const CardGridPage = () => (
  <div className="flex flex-col gap-8 p-4">
    <Page crumbs={[{ label: 'Apps' }]} actions={['Sync', 'Create app']}>
      <PlaceholderGrid rows={9} columns={3} height="h-[8rem]" />
    </Page>
  </div>
)

export const TabbedPage = () => (
  <div className="flex flex-col gap-8 p-4">
    <Page
      tabs={appTabs('app-01', 'br-main')}
      crumbs={[{ label: 'Apps', path: '/apps' }, { label: 'app-01' }]}
      actions={['Sync']}
    >
      <PlaceholderGrid rows={6} columns={2} />
    </Page>
  </div>
)

export const DocumentPage = () => (
  <div className="flex flex-col gap-8 p-4">
    <Page crumbs={[{ label: 'Docs' }]}>
      <PlaceholderGrid rows={3} height="h-[10rem]" />
    </Page>
  </div>
)
