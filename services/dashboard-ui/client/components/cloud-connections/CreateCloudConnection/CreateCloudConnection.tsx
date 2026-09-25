import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormSelect } from '@/components/common/form/FormSelect'
import { FormTextarea } from '@/components/common/form/FormTextarea'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TCloudConnection } from '@/types'
import {
  createCloudConnectionSchema,
  type CreateCloudConnectionValues,
} from './schema'

const CLOUD_OPTIONS = [
  { value: 'aws', label: 'Amazon Web Services' },
  {
    value: 'azure',
    label: 'Microsoft Azure',
  },
  {
    value: 'gcp',
    label: 'Google Cloud',
  },
]

const CAPABILITY_OPTIONS = [
  { value: 'stacks', label: 'Install stacks' },
  { value: 'images', label: 'Pull images' },
  { value: 'both', label: 'Both' },
]

interface ICreateCloudConnectionModal extends Omit<IModal, 'onSubmit'> {
  defaultPlatform?: 'aws' | 'azure' | 'gcp'
  connection: TCloudConnection | null
  error: TAPIError | null
  isPending: boolean
  isVerifying: boolean
  verifyError: TAPIError | null
  onSubmit: (values: CreateCloudConnectionValues) => void
  onVerify: () => void
  onDone: () => void
}

export const CreateCloudConnectionModal = ({
  defaultPlatform = 'aws',
  connection,
  error,
  isPending,
  isVerifying,
  verifyError,
  onSubmit,
  onVerify,
  onDone,
  ...props
}: ICreateCloudConnectionModal) => {
  const form = useForm({
    defaultValues: {
      name: '',
      platform: defaultPlatform,
      targetId: '',
      tenantId: '',
      capabilitySet: defaultPlatform === 'gcp' ? 'images' : 'both',
      registry: '',
      repositories: '',
      principal: '',
      identityProvider: '',
    } as CreateCloudConnectionValues,
    validators: {
      onMount: createCloudConnectionSchema,
      onChange: createCloudConnectionSchema,
    },
    onSubmit: ({ value }) => onSubmit(value),
  })
  const canSubmit = useStore(form.store, (state) => state.canSubmit)
  const platform = useStore(form.store, (state) => state.values.platform)
  const capabilitySet = useStore(
    form.store,
    (state) => state.values.capabilitySet
  )

  if (connection) {
    const verified = connection.status === 'verified'
    return (
      <Modal
        heading="Set up cloud connection"
        size="lg"
        secondaryActionTrigger={{ children: 'Done', onClick: onDone }}
        primaryActionTrigger={{
          children: isVerifying ? 'Verifying…' : 'Verify connection',
          disabled: isVerifying,
          onClick: onVerify,
          variant: 'primary',
        }}
        {...props}
      >
        <FormErrorBanner
          error={verifyError}
          fallback="Unable to verify connection"
        />
        {verified ? (
          <Banner theme="success">
            Connection verified with {connection.capabilities?.join(' and ')}{' '}
            capabilities.
          </Banner>
        ) : connection.status_message ? (
          <Banner theme={connection.status === 'error' ? 'error' : 'info'}>
            {connection.status_message}
          </Banner>
        ) : null}
        <div className="grid gap-4 rounded-lg border p-4">
          <div className="flex flex-col gap-1">
            <Text variant="label" theme="neutral">
              Issuer
            </Text>
            <Text family="mono" className="break-all">
              {connection.setup.issuer_url}
            </Text>
          </div>
          <div className="flex flex-col gap-1">
            <Text variant="label" theme="neutral">
              Subject
            </Text>
            <Text family="mono" className="break-all">
              {connection.setup.subject}
            </Text>
          </div>
        </div>
        <Text theme="neutral">
          Apply one of these configurations in the{' '}
          {connection.platform === 'azure'
            ? 'Azure subscription'
            : connection.platform === 'gcp'
              ? 'GCP project'
              : 'AWS account'}
          , then verify the connection.
        </Text>
        {connection.platform === 'azure' ? (
          <Tabs
            tabs={{
              terraform: (
                <CodeBlock language="hcl" showCopy>
                  {connection.setup.terraform}
                </CodeBlock>
              ),
              cli: (
                <CodeBlock language="bash" showCopy>
                  {connection.setup.cli}
                </CodeBlock>
              ),
              portal: (
                <CodeBlock language="json" showCopy>
                  {connection.setup.portal_json ?? ''}
                </CodeBlock>
              ),
            }}
            tabLabels={{ cli: 'Azure CLI', portal: 'Portal JSON' }}
          />
        ) : connection.platform === 'gcp' ? (
          <Tabs
            tabs={{
              terraform: (
                <CodeBlock language="hcl" showCopy>
                  {connection.setup.terraform}
                </CodeBlock>
              ),
              cli: (
                <CodeBlock language="bash" showCopy>
                  {connection.setup.cli}
                </CodeBlock>
              ),
            }}
            tabLabels={{ cli: 'gcloud CLI' }}
          />
        ) : (
          <Tabs
            tabs={{
              terraform: (
                <CodeBlock language="hcl" showCopy>
                  {connection.setup.terraform}
                </CodeBlock>
              ),
              cli: (
                <CodeBlock language="bash" showCopy>
                  {connection.setup.cli}
                </CodeBlock>
              ),
              cloudformation: (
                <CodeBlock language="yaml" showCopy>
                  {connection.setup.cloudformation}
                </CodeBlock>
              ),
            }}
            tabLabels={{ cli: 'AWS CLI', cloudformation: 'CloudFormation' }}
          />
        )}
      </Modal>
    )
  }

  return (
    <Modal
      heading="Create cloud connection"
      size="lg"
      primaryActionTrigger={{
        children: isPending ? 'Creating…' : 'Create connection',
        disabled: !canSubmit || isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(event) => event.preventDefault()}
        className="grid gap-6 sm:grid-cols-2"
      >
        <div className="sm:col-span-2">
          <FormErrorBanner
            error={error}
            fallback="Unable to create cloud connection"
          />
        </div>
        <form.Field name="platform">
          {(field) => (
            <FormSelect
              field={field}
              options={CLOUD_OPTIONS}
              labelProps={{ labelText: 'Cloud' }}
              disabled={isPending}
            />
          )}
        </form.Field>
        {platform === 'azure' && (
          <form.Field name="tenantId">
            {(field) => (
              <FormInput
                field={field}
                labelProps={{ labelText: 'Entra tenant ID' }}
                placeholder="00000000-0000-4000-8000-000000000000"
                disabled={isPending}
              />
            )}
          </form.Field>
        )}
        <form.Field name="targetId">
          {(field) => (
            <FormInput
              field={field}
              labelProps={{
                labelText:
                  platform === 'azure'
                    ? 'Azure subscription ID'
                    : platform === 'gcp'
                      ? 'GCP project ID'
                      : 'AWS account ID',
              }}
              placeholder={
                platform === 'azure'
                  ? '00000000-0000-4000-8000-000000000000'
                  : platform === 'gcp'
                    ? 'acme-production'
                    : '123456789012'
              }
              disabled={isPending}
            />
          )}
        </form.Field>
        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              labelProps={{ labelText: 'Name' }}
              placeholder={
                platform === 'azure'
                  ? 'Production Azure'
                  : platform === 'gcp'
                    ? 'Production GCP'
                    : 'Production AWS'
              }
              disabled={isPending}
            />
          )}
        </form.Field>
        <form.Field name="capabilitySet">
          {(field) => (
            <FormSelect
              field={field}
              options={
                platform === 'gcp'
                  ? [{ value: 'images', label: 'Pull images' }]
                  : CAPABILITY_OPTIONS
              }
              labelProps={{ labelText: 'Access' }}
              disabled={isPending}
            />
          )}
        </form.Field>
        {platform === 'azure' && capabilitySet !== 'stacks' && (
          <form.Field name="registry">
            {(field) => (
              <FormInput
                field={field}
                labelProps={{ labelText: 'Registry name' }}
                placeholder="acmecontainers"
                disabled={isPending}
              />
            )}
          </form.Field>
        )}
        {capabilitySet !== 'stacks' && (
          <div className="sm:col-span-2">
            <form.Field name="repositories">
              {(field) => (
                <FormTextarea
                  field={field}
                  labelProps={{
                    labelText: `${platform === 'azure' ? 'ACR' : platform === 'gcp' ? 'Artifact Registry' : 'ECR'} repositories (optional)`,
                  }}
                  helperText={
                    platform === 'gcp'
                      ? 'One location/repository per line. Used to scope image access.'
                      : 'One repository name per line. Used to scope image access.'
                  }
                  placeholder={
                    platform === 'gcp'
                      ? 'us-central1/backend\nus-central1/worker'
                      : 'backend\nworker'
                  }
                  disabled={isPending}
                />
              )}
            </form.Field>
          </div>
        )}
        <div className="sm:col-span-2">
          <form.Field name="principal">
            {(field) => (
              <FormInput
                field={field}
                labelProps={{
                  labelText:
                    platform === 'azure'
                      ? 'Client ID'
                      : platform === 'gcp'
                        ? 'Service account email'
                        : 'IAM role ARN',
                }}
                placeholder={
                  platform === 'azure'
                    ? '00000000-0000-4000-8000-000000000000'
                    : platform === 'gcp'
                      ? 'nuon-cloud-connection@acme-production.iam.gserviceaccount.com'
                      : 'arn:aws:iam::123456789012:role/nuon-cloud-connection'
                }
                disabled={isPending}
              />
            )}
          </form.Field>
        </div>
        {platform === 'gcp' && (
          <div className="sm:col-span-2">
            <form.Field name="identityProvider">
              {(field) => (
                <FormInput
                  field={field}
                  labelProps={{ labelText: 'Workload Identity Provider' }}
                  placeholder="projects/123456789/locations/global/workloadIdentityPools/nuon/providers/connection"
                  disabled={isPending}
                />
              )}
            </form.Field>
          </div>
        )}
      </form>
    </Modal>
  )
}
