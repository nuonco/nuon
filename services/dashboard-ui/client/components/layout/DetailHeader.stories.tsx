export default {
  title: 'Layout/DetailHeader',
}

import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DetailHeader } from './DetailHeader'
import { PageSection } from './PageSection'

const created = '2026-08-24T10:12:00Z'
const updated = '2026-08-24T10:19:42Z'

const runMetadata = (
  <>
    <LabeledStatus label="Status" statusProps={{ status: 'active' }} />
    <LabeledValue label="Duration">
      <Duration variant="subtext" beginTime={created} endTime={updated} />
    </LabeledValue>
    <LabeledValue label="Install">
      <Link href="#">acme-payments</Link>
    </LabeledValue>
    <LabeledValue label="Config">
      <Link href="#">api</Link>
    </LabeledValue>
    <LabeledValue label="Execution role">
      <Text variant="subtext" family="mono">
        nuon-deploy
      </Text>
    </LabeledValue>
  </>
)

export const Run = () => (
  <PageSection>
    <DetailHeader
      title="api deploy"
      icon={<Icon variant="CubeIcon" size={18} />}
      id="dep01hzk8t3fqp2r9x4m7wcn5vb"
      identity={
        <Time time={created} format="relative" variant="subtext" theme="info" />
      }
      actions={
        <>
          <Button variant="secondary">Deploy history</Button>
          <Button variant="secondary">Manage</Button>
        </>
      }
      metadata={runMetadata}
    />
  </PageSection>
)

export const RunWithFooterAction = () => (
  <PageSection>
    <DetailHeader
      title="api deploy"
      id="dep01hzk8t3fqp2r9x4m7wcn5vb"
      metadata={runMetadata}
    >
      <Button href="#">View workflow</Button>
    </DetailHeader>
  </PageSection>
)

export const Entity = () => (
  <PageSection>
    <DetailHeader
      backLink={false}
      title="Sandbox details"
      id="sbx01hzk8t3fqp2r9x4m7wcn5vb"
      status={<Badge theme="info">Terraform</Badge>}
      actions={<Button variant="secondary">Manage</Button>}
    />
  </PageSection>
)

export const WithDescription = () => (
  <PageSection>
    <DetailHeader
      title="acme-payments connection"
      description="Connected to the acme-payments GitHub organization."
      id="vcs01hzk8t3fqp2r9x4m7wcn5vb"
      actions={<Button variant="danger">Remove connection</Button>}
    />
  </PageSection>
)

export const Loading = () => (
  <PageSection>
    <DetailHeader
      title="api deploy"
      id="dep01hzk8t3fqp2r9x4m7wcn5vb"
      loading
      metadata={
        <>
          <LabeledStatus label="Status" loading />
          <LabeledValue label="Duration" loading />
          <LabeledValue label="Install" loading />
        </>
      }
    />
  </PageSection>
)
