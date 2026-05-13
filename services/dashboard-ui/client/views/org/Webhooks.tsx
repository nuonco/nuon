import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import {
  CreateWebhookButton,
  PayloadFieldReference,
  SamplePayload,
  WebhooksTable,
} from '@/components/webhooks'
import { useOrg } from '@/hooks/use-org'

export const Webhooks = () => {
  const { org } = useOrg()

  return (
    <PageLayout className="pb-6">
      <PageTitle title={`Webhooks | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/webhooks`, text: 'Webhooks' },
        ]}
      />
      <PageHeader className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Webhooks
          </Text>
          <Text theme="neutral">
            Receive workflow and workflow step lifecycle events for this org
            as CloudEvents v1.0 payloads.
          </Text>
        </HeadingGroup>
        <CreateWebhookButton />
      </PageHeader>
      <PageContent>
        <PageSection>
          <div className="flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Active webhooks
            </Text>
            <WebhooksTable shouldPoll />
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="base" weight="strong">
              Payload format
            </Text>
            <Text variant="body" theme="neutral">
              Webhooks deliver CloudEvents v1.0 JSON payloads of type{' '}
              <span className="font-mono">com.nuon.workflow.lifecycle.v1</span>{' '}
              and{' '}
              <span className="font-mono">com.nuon.workflow_step.lifecycle.v1</span>.
              When a signing secret is set, requests are signed with HMAC-SHA256
              and the hex-encoded signature is sent in the{' '}
              <span className="font-mono">X-Nuon-Signature</span> header.{' '}
              <a
                href="https://docs.nuon.co/webhooks"
                target="_blank"
                rel="noreferrer noopener"
                className="text-primary-600 dark:text-primary-400 hover:underline inline-flex items-center gap-1"
              >
                Read the docs
                <Icon variant="ArrowSquareOutIcon" size={14} />
              </a>
            </Text>
            <SamplePayload />
          </div>

          <PayloadFieldReference />
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}
