import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export type ClaimCondition = { key: string; value: string }

export type OIDCTrustPolicyFormInput = {
  name: string
  issuerUrl: string
  audience: string
  role: string
  tokenDurationSeconds: string
  claimConditions: ClaimCondition[]
}

const ROLE_OPTIONS = [
  { value: 'org_read_only', label: 'org_read_only' },
  { value: 'org_builder', label: 'org_builder' },
  { value: 'org_support', label: 'org_support' },
  { value: 'org_admin', label: 'org_admin' },
]

export const hasSubCondition = (claimConditions: ClaimCondition[]) =>
  claimConditions.some(
    (condition) => condition.key.trim() === 'sub' && condition.value.trim()
  )

export const CreateOIDCTrustPolicyModal = ({
  isPending,
  error,
  onSubmit,
  initialValues,
  lockIssuer,
  reservedNames,
  ...props
}: {
  isPending: boolean
  error: TAPIError | null
  onSubmit: (input: OIDCTrustPolicyFormInput) => void
  initialValues?: Partial<OIDCTrustPolicyFormInput>
  lockIssuer?: boolean
  reservedNames?: string[]
} & Omit<IModal, 'onSubmit'>) => {
  const [name, setName] = useState(initialValues?.name ?? '')
  const [issuerUrl, setIssuerUrl] = useState(initialValues?.issuerUrl ?? '')
  const [audience, setAudience] = useState(initialValues?.audience ?? '')
  const [role, setRole] = useState(initialValues?.role ?? 'org_read_only')
  const [tokenDurationSeconds, setTokenDurationSeconds] = useState(
    initialValues?.tokenDurationSeconds ?? ''
  )
  const [claimConditions, setClaimConditions] = useState<ClaimCondition[]>(
    initialValues?.claimConditions ?? [{ key: 'sub', value: '' }]
  )

  const trimmedName = name.trim()
  const trimmedIssuerUrl = issuerUrl.trim()
  const trimmedAudience = audience.trim()
  const isValidIssuerUrl = /^https?:\/\/.+/i.test(trimmedIssuerUrl)
  const isNameTaken = !!reservedNames?.some(
    (reserved) => reserved.trim().toLowerCase() === trimmedName.toLowerCase()
  )
  const canSubmit =
    !isPending &&
    !!trimmedName &&
    !isNameTaken &&
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
          Create OIDC trust policy
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Creating...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="PlusIcon" />
            Create trust policy
          </span>
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
          }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to create trust policy'}
          </Banner>
        ) : null}

        <Text variant="body" theme="neutral">
          Trust policies let a CI/CD provider exchange an OIDC token for
          short-lived org access, without storing a static API token.
        </Text>

        <div className="flex flex-col gap-2">
          <Label htmlFor="policy-name">Name</Label>
          <Input
            id="policy-name"
            placeholder="GitHub Actions CI"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          {isNameTaken ? (
            <Text variant="subtext" theme="error">
              A trust policy named {trimmedName} already exists. Choose a
              different name.
            </Text>
          ) : null}
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="policy-issuer-url">Issuer URL</Label>
          <Input
            id="policy-issuer-url"
            placeholder="https://token.actions.githubusercontent.com"
            type="url"
            value={issuerUrl}
            onChange={(e) => setIssuerUrl(e.target.value)}
            required
            readOnly={lockIssuer}
            disabled={lockIssuer}
          />
          <Text variant="subtext" theme="neutral">
            Must be an absolute http or https URL.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="policy-audience">Audience</Label>
          <Input
            id="policy-audience"
            placeholder="https://api.nuon.co"
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
          options={ROLE_OPTIONS}
          value={role}
          onChange={(e) => setRole(e.target.value)}
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
            Defaults to 3600. Maximum is 86400.
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
