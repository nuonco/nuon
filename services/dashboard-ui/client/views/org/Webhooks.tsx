import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { ListPage } from '@/components/layout/ListPage'
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
    <>
      <PageTitle title="Webhooks" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/settings`, text: 'Settings' },
          { path: `/${org.id}/settings/webhooks`, text: 'Webhooks' },
        ]}
      />
      <ListPage
        title="Webhooks"
        description="Receive workflow and workflow step lifecycle events for this org as CloudEvents v1.0 payloads."
        createAction={<CreateWebhookButton />}
      >
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
            <span className="font-mono">
              com.nuon.workflow_step.lifecycle.v1
            </span>
            . When a signing secret is set, requests are signed with HMAC-SHA256
            and the hex-encoded signature is sent in the{' '}
            <span className="font-mono">X-Nuon-Signature</span> header.{' '}
            <Link
              href="https://docs.nuon.co/webhooks"
              isExternal
              variant="inline"
            >
              Read the docs
            </Link>
          </Text>
          <SamplePayload />
        </div>

        <PayloadFieldReference />
      </ListPage>
    </>
  )
}
