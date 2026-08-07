import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Toggle } from '@/components/common/form/Toggle'
import {
  hasSubCondition,
  type ClaimCondition,
} from '@/components/oidc-trust-policies/CreateOIDCTrustPolicy'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TOIDCTrustPolicy } from '@/types'

export type EditOIDCTrustPolicyFormInput = {
  name: string
  issuerUrl: string
  audience: string
  role: string
  tokenDurationSeconds: string
  claimConditions: ClaimCondition[]
  enabled: boolean
}

const conditionsToRows = (
  claimConditions: TOIDCTrustPolicy['claim_conditions']
): ClaimCondition[] => {
  const entries = Object.entries(claimConditions ?? {})
  return entries.length
    ? entries.map(([key, value]) => ({ key, value }))
    : [{ key: 'sub', value: '' }]
}

export const EditOIDCTrustPolicyModal = ({
  policy,
  isPending,
  error,
  roleOptions,
  onSubmit,
  ...props
}: {
  policy: TOIDCTrustPolicy
  isPending: boolean
  error: TAPIError | null
  roleOptions: { value: string; label: string }[]
  onSubmit: (input: EditOIDCTrustPolicyFormInput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [name, setName] = useState(policy.name ?? '')
  const [issuerUrl, setIssuerUrl] = useState(policy.issuer_url ?? '')
  const [audience, setAudience] = useState(policy.audience ?? '')
  const [role, setRole] = useState(policy.role ?? 'org_read_only')
  const [tokenDurationSeconds, setTokenDurationSeconds] = useState(
    policy.token_duration_seconds ? String(policy.token_duration_seconds) : ''
  )
  const [enabled, setEnabled] = useState(policy.enabled ?? true)
  const [claimConditions, setClaimConditions] = useState<ClaimCondition[]>(() =>
    conditionsToRows(policy.claim_conditions)
  )

  const trimmedName = name.trim()
  const trimmedIssuerUrl = issuerUrl.trim()
  const trimmedAudience = audience.trim()
  const isValidIssuerUrl = /^https?:\/\/.+/i.test(trimmedIssuerUrl)
  const canSubmit =
    !isPending &&
    !!trimmedName &&
    isValidIssuerUrl &&
    !!trimmedAudience &&
    hasSubCondition(claimConditions)

  const updateClaimCondition = (
    index: number,
    field: 'key' | 'value',
    value: string
  ) => {
    setClaimConditions((prev) =>
      prev.map((condition, i) =>
        i === index ? { ...condition, [field]: value } : condition
      )
    )
  }

  const addClaimCondition = () =>
    setClaimConditions((prev) => [...prev, { key: '', value: '' }])

  const removeClaimCondition = (index: number) =>
    setClaimConditions((prev) => prev.filter((_, i) => i !== index))

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="ShieldCheckIcon" size="24" />
          Edit trust policy
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving...
          </span>
        ) : (
          'Save changes'
        ),
        disabled: !canSubmit,
        onClick: () =>
          onSubmit({
            name: trimmedName,
            issuerUrl: trimmedIssuerUrl,
            audience: trimmedAudience,
            role,
            tokenDurationSeconds,
            claimConditions,
            enabled,
          }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to update trust policy'}
          </Banner>
        ) : null}

        <Toggle
          checked={enabled}
          onChange={setEnabled}
          label="Enabled"
          description="Disabled policies reject token exchange requests."
        />

        <div className="flex flex-col gap-2">
          <Label htmlFor="policy-name">Name</Label>
          <Input
            id="policy-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="policy-issuer-url">Issuer URL</Label>
          <Input
            id="policy-issuer-url"
            type="url"
            value={issuerUrl}
            onChange={(e) => setIssuerUrl(e.target.value)}
            required
          />
          <Text variant="subtext" theme="neutral">
            Must be an absolute http or https URL.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="policy-audience">Audience</Label>
          <Input
            id="policy-audience"
            value={audience}
            onChange={(e) => setAudience(e.target.value)}
            required
          />
          <Text variant="subtext" theme="neutral">
            The expected `aud` claim value on the presented token.
          </Text>
        </div>

        <Select
          labelProps={{ labelText: 'Role' }}
          options={roleOptions}
          value={role}
          onChange={(value) => setRole(value)}
          helperText="Org role granted to tokens exchanged with this policy."
        />

        <div className="flex flex-col gap-2">
          <Label htmlFor="policy-token-duration">
            Token duration in seconds (optional)
          </Label>
          <Input
            id="policy-token-duration"
            placeholder="3600"
            type="number"
            min={1}
            max={86400}
            value={tokenDurationSeconds}
            onChange={(e) => setTokenDurationSeconds(e.target.value)}
          />
          <Text variant="subtext" theme="neutral">
            Maximum is 86400.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label>Claim conditions</Label>
          <Text variant="subtext" theme="neutral">
            All conditions must match the presented token. A `sub` condition is
            required.
          </Text>
          <div className="flex flex-col gap-2">
            {claimConditions.map((condition, index) => (
              <div key={index} className="flex items-center gap-2">
                <Input
                  placeholder="sub"
                  value={condition.key}
                  onChange={(e) =>
                    updateClaimCondition(index, 'key', e.target.value)
                  }
                />
                <Input
                  placeholder="repo:acme/app:ref:refs/heads/main"
                  value={condition.value}
                  onChange={(e) =>
                    updateClaimCondition(index, 'value', e.target.value)
                  }
                />
                <Button
                  variant="icon"
                  aria-label="Remove claim condition"
                  disabled={claimConditions.length === 1}
                  onClick={() => removeClaimCondition(index)}
                >
                  <Icon variant="TrashIcon" size={14} />
                </Button>
              </div>
            ))}
          </div>
          <Button
            variant="secondary"
            size="sm"
            className="w-fit"
            onClick={addClaimCondition}
          >
            <Icon variant="PlusIcon" size={14} />
            Add condition
          </Button>
        </div>
      </div>
    </Modal>
  )
}
