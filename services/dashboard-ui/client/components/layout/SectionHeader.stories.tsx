export default {
  title: 'Layout/SectionHeader',
}

import { Button } from '@/components/common/Button'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { PageSection } from './PageSection'
import { SectionHeader } from './SectionHeader'

export const Page = () => (
  <SectionHeader
    variant="page"
    title="Installs"
    description="View and manage all deployed installs here."
  />
)

export const PageWithActions = () => (
  <SectionHeader
    variant="page"
    title="Webhooks"
    description="Receive workflow lifecycle events for this org as CloudEvents v1.0 payloads."
    actions={<Button variant="primary">Create webhook</Button>}
  />
)

export const PageWithStatus = () => (
  <SectionHeader
    variant="page"
    title="acme-payments connection"
    description="Connected 3 days ago."
    status={<Status status="active" variant="badge" />}
    actions={
      <>
        <Button variant="secondary">Add connection</Button>
        <Button variant="danger">Remove connection</Button>
      </>
    }
  />
)

export const Section = () => (
  <PageSection>
    <SectionHeader
      title="App components"
      description="Manage the components that make up your application."
    />
    <Text theme="neutral">Section body</Text>
  </PageSection>
)

export const SectionWithActions = () => (
  <PageSection>
    <SectionHeader
      title="App components"
      description="Manage the components that make up your application."
      actions={<Button variant="primary">Create component</Button>}
    />
    <Text theme="neutral">Section body</Text>
  </PageSection>
)

export const TitleOnly = () => (
  <PageSection>
    <SectionHeader title="Active webhooks" />
    <Text theme="neutral">Section body</Text>
  </PageSection>
)
