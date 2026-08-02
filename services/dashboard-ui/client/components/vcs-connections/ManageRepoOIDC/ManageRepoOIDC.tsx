import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Toggle } from '@/components/common/form/Toggle'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TOIDCTrustPolicy } from '@/types'

export const ManageRepoOIDCModal = ({
  policies,
  isLoading,
  error,
  togglingId,
  onToggle,
  createSlot,
  renderDelete,
  ...props
}: {
  policies: TOIDCTrustPolicy[]
  isLoading: boolean
  error?: TAPIError | null
  togglingId?: string
  onToggle: (policy: TOIDCTrustPolicy, next: boolean) => void
  createSlot: ReactNode
  renderDelete: (policy: TOIDCTrustPolicy) => ReactNode
} & Omit<IModal, 'onSubmit'>) => (
  <Modal
    heading={
      <Text flex className="gap-4" variant="h3" weight="strong">
        <Icon variant="ShieldCheckIcon" size="24" />
        Manage OIDC
      </Text>
    }
    showFooter={false}
    {...props}
  >
    <div className="flex flex-col gap-6">
      {error ? (
        <Banner theme="error">
          {error?.error || 'Unable to load trust policies'}
        </Banner>
      ) : null}

      {isLoading ? (
        <div className="flex flex-col gap-2">
          {Array.from({ length: 2 }).map((_, idx) => (
            <Skeleton key={idx} height="56px" />
          ))}
        </div>
      ) : policies.length ? (
        <div className="flex flex-col gap-2">
          {policies.map((policy) => (
            <div
              key={policy.id}
              className="flex items-center justify-between gap-4 py-2 px-4 border rounded-md"
            >
              <div className="flex flex-col gap-1">
                <Text variant="subtext" weight="strong">
                  {policy.name}
                </Text>
                <Text
                  flex
                  className="gap-2 flex-wrap"
                  variant="subtext"
                  theme="neutral"
                >
                  <Badge size="sm" variant="code">
                    {policy.role}
                  </Badge>
                  {policy.claim_conditions?.sub}
                </Text>
              </div>
              <div className="flex items-center gap-3">
                <Toggle
                  checked={!!policy.enabled}
                  disabled={togglingId === policy.id}
                  label={policy.enabled ? 'Enabled' : 'Disabled'}
                  onChange={(next) => onToggle(policy, next)}
                />
                {renderDelete(policy)}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <EmptyState
          variant="search"
          size="sm"
          emptyTitle="No trust policy yet"
          emptyMessage="Create one to let this repository authenticate without a stored token."
        />
      )}

      <div>{createSlot}</div>
    </div>
  </Modal>
)
