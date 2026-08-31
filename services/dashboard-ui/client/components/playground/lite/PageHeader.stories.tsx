import { PageHeader } from './PageHeader'

export default {
  title: 'Playground/Lite/PageHeader',
}

export const Default = () => <PageHeader crumbs={[{ label: 'Apps' }]} />

export const DeepCrumbs = () => (
  <PageHeader
    crumbs={[
      { label: 'Installs', path: '/installs' },
      { label: 'install-02', path: '/installs/install-02' },
      { label: 'Inputs' },
    ]}
  />
)
