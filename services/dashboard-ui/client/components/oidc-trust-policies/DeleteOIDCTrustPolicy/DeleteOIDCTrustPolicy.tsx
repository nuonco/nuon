import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteOIDCTrustPolicyModal = ({
  policyName,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  policyName: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
} & IModal) => {
  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="WarningIcon" size="24" />
          Delete trust policy?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Deleting...
          </span>
        ) : (
          'Delete trust policy'
        ),
        disabled: isPending,
        onClick: () => onSubmit(),
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to delete trust policy'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-3">
          <Text variant="base" weight="strong">
            Deleting {policyName} will stop it from exchanging OIDC tokens for
            org access.
          </Text>
          <Text variant="body" theme="neutral">
            <strong>Warning:</strong> any outstanding federated tokens issued
            through this policy will be revoked, since the backing service
            account is removed.
          </Text>
        </div>
      </div>
    </Modal>
  )
}
