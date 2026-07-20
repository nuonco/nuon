import { useRef, type FormEvent } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAWSAccountConnection, TAPIError } from '@/types'

interface IConnectAWSAccountModal extends Omit<IModal, 'onSubmit'> {
  connection?: TAWSAccountConnection
  error?: TAPIError | null
  isLoading?: boolean
  isPending: boolean
  isVerifying?: boolean
  onCreate: (values: {
    name: string
    accountId: string
    region: string
  }) => void
  onSetRole: (roleArn: string) => void
  onVerify: () => void
}

export const ConnectAWSAccountModal = ({
  connection,
  error,
  isLoading,
  isPending,
  isVerifying,
  onCreate,
  onSetRole,
  onVerify,
  ...props
}: IConnectAWSAccountModal) => {
  const formRef = useRef<HTMLFormElement>(null)
  const hasRole = !!connection?.role_arn

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const values = Object.fromEntries(new FormData(event.currentTarget))
    if (connection) {
      onSetRole(values.role_arn as string)
      return
    }
    onCreate({
      name: values.name as string,
      accountId: values.account_id as string,
      region: values.default_region as string,
    })
  }

  const actionLabel = connection ? 'Save role ARN' : 'Create AWS connection'

  return (
    <Modal
      heading={
        <div className="flex items-center gap-3">
          <Icon variant="AWS" />
          <Text variant="h3" weight="strong">
            Connect AWS account
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isPending ? 'Connecting...' : actionLabel,
        disabled: isPending,
        onClick: () => formRef.current?.requestSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to connect the AWS account.'}
          </Banner>
        ) : null}

        {isLoading ? (
          <Text>Loading AWS account connection...</Text>
        ) : !connection ? (
          <form
            ref={formRef}
            onSubmit={handleSubmit}
            className="flex flex-col gap-4"
          >
            <Input
              name="name"
              required
              labelProps={{ labelText: 'Connection name' }}
              placeholder="Demo smoke tests"
            />
            <Input
              name="account_id"
              required
              labelProps={{ labelText: 'AWS account ID' }}
              pattern="[0-9]{12}"
              placeholder="123456789012"
            />
            <Input
              name="default_region"
              required
              defaultValue="us-west-2"
              labelProps={{ labelText: 'Default region' }}
            />
          </form>
        ) : (
          <>
            <Text>
              Configure <strong>{connection.name}</strong> in AWS with this
              trust policy, then enter the role ARN.
            </Text>
            <div className="flex items-center gap-2">
              <Status status={connection.verification_status} />
              {connection.last_checked_at ? (
                <Text variant="subtext" theme="neutral">
                  Last checked{' '}
                  <Time
                    time={connection.last_checked_at}
                    format="relative"
                    variant="subtext"
                    shouldTick
                  />
                </Text>
              ) : null}
            </div>
            <CodeBlock language="json" showCopy>
              {JSON.stringify(connection.trust_policy, null, 2)}
            </CodeBlock>
            <Text variant="subtext">
              External ID: <code>{connection.external_id}</code>
            </Text>
            <form
              ref={formRef}
              onSubmit={handleSubmit}
              className="flex flex-col gap-4"
            >
              <Input
                name="role_arn"
                required
                defaultValue={connection.role_arn}
                labelProps={{ labelText: 'Customer role ARN' }}
                placeholder={`arn:aws:iam::${connection.account_id}:role/nuon-smoke-tests`}
              />
            </form>
            {hasRole ? (
              <Button
                disabled={isPending}
                onClick={onVerify}
                variant="secondary"
              >
                <Icon
                  variant={isVerifying ? 'Loading' : 'ArrowClockwiseIcon'}
                />
                {isVerifying ? 'Checking connection...' : 'Re-check connection'}
              </Button>
            ) : null}
            {connection.verification_status === 'error' ? (
              <Banner theme="error">
                {connection.verification_message ||
                  'AWS connection verification failed.'}
              </Banner>
            ) : null}
          </>
        )}
      </div>
    </Modal>
  )
}
