export default {
  title: 'Layout/PageHeader',
}

import { PageHeader } from './PageHeader'
import { PageHeadingGroup } from './PageHeadingGroup'
import { Button } from '@/components/common/Button'

export const Default = () => (
  <PageHeader>
    <PageHeadingGroup title="Page title" subtitle="A short description of the page" />
  </PageHeader>
)

export const WithActions = () => (
  <PageHeader>
    <PageHeadingGroup title="Installs" subtitle="Manage your installs" />
    <Button variant="primary">Create install</Button>
  </PageHeader>
)
