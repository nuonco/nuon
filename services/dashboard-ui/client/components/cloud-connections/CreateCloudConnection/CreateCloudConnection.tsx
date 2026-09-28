import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormSelect } from '@/components/common/form/FormSelect'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TCloudConnection } from '@/types'
import {
  createCloudConnectionSchema,
  type CreateCloudConnectionValues,
} from './schema'

const PRESET_OPTIONS = [
  { value: 'stacks', label: 'Install stacks' },
  { value: 'custom', label: 'Custom permissions' },
]

interface ICreateCloudConnectionModal extends Omit<IModal, 'onSubmit'> {
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
      targetId: '',
      preset: 'stacks',
      principal: '',
    } as CreateCloudConnectionValues,
    validators: {
      onMount: createCloudConnectionSchema,
      onChange: createCloudConnectionSchema,
    },
    onSubmit: ({ value }) => onSubmit(value),
  })
  const canSubmit = useStore(form.store, (state) => state.canSubmit)
  const preset = useStore(form.store, (state) => state.values.preset)

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
            Connection verified with the {connection.preset} preset.
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
          Apply one of these configurations in the AWS account, then verify the
          connection.
          {connection.preset === 'custom' &&
            ' Attach your own permissions policy to the role; verification checks identity only.'}
        </Text>
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
        <form.Field name="targetId">
          {(field) => (
            <FormInput
              field={field}
              labelProps={{ labelText: 'AWS account ID' }}
              placeholder="123456789012"
              disabled={isPending}
            />
          )}
        </form.Field>
        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              labelProps={{ labelText: 'Name' }}
              placeholder="Production AWS"
              disabled={isPending}
            />
          )}
        </form.Field>
        <form.Field name="preset">
          {(field) => (
            <FormSelect
              field={field}
              options={PRESET_OPTIONS}
              labelProps={{ labelText: 'Access preset' }}
              disabled={isPending}
            />
          )}
        </form.Field>
        {preset === 'custom' && (
          <Text theme="neutral">
            Attach your own permissions policy to the role. Verification checks
            identity only.
          </Text>
        )}
        <div className="sm:col-span-2">
          <form.Field name="principal">
            {(field) => (
              <FormInput
                field={field}
                labelProps={{ labelText: 'IAM role ARN' }}
                placeholder="arn:aws:iam::123456789012:role/nuon-cloud-connection"
                disabled={isPending}
              />
            )}
          </form.Field>
        </div>
      </form>
    </Modal>
  )
}
