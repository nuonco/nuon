import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TVCSConnectionRepo } from '@/types'

export type ClaimCondition = { key: string; value: string }

export type OIDCPreset = 'github_actions' | 'custom'

export type OIDCTrustPolicyFormInput = {
  name: string
  issuerUrl: string
  audience: string
  role: string
  tokenDurationSeconds: string
  claimConditions: ClaimCondition[]
}

export const GITHUB_ACTIONS_ISSUER =
  'https://token.actions.githubusercontent.com'

const PRESET_OPTIONS = [
  { value: 'github_actions', label: 'GitHub Actions' },
  { value: 'custom', label: 'Custom' },
]

export const hasSubCondition = (claimConditions: ClaimCondition[]) =>
  claimConditions.some(
    (condition) => condition.key.trim() === 'sub' && condition.value.trim()
  )

export const githubSubClaim = (repoFullName: string, branch: string) =>
  `repo:${repoFullName}:ref:refs/heads/${branch}`

export const defaultRepoPolicyName = (
  repoFullName: string,
  reservedNames: string[] = []
) => {
  const taken = new Set(
    reservedNames.map((reserved) => reserved.trim().toLowerCase())
  )
  const baseName = `github-${repoFullName.split('/').pop() ?? repoFullName}`
  let name = baseName
  for (let n = 2; taken.has(name.toLowerCase()); n++) {
    name = `${baseName}-${n}`
  }
  return name
}

export const CreateOIDCTrustPolicyModal = ({
  isPending,
  error,
  onSubmit,
  repos,
  isLoadingRepos,
  hasVCSConnections,
  vcsConnectionsHref,
  githubAudience,
  initialRepoFullName,
  initialRepoDefaultBranch,
  lockPreset,
  reservedNames,
  roleOptions,
  ...props
}: {
  isPending: boolean
  error: TAPIError | null
  onSubmit: (input: OIDCTrustPolicyFormInput) => void
  repos: TVCSConnectionRepo[]
  isLoadingRepos?: boolean
  hasVCSConnections?: boolean
  vcsConnectionsHref: string
  githubAudience: string
  initialRepoFullName?: string
  initialRepoDefaultBranch?: string
  lockPreset?: boolean
  reservedNames?: string[]
  roleOptions: { value: string; label: string; description?: string }[]
} & Omit<IModal, 'onSubmit'>) => {
  const [preset, setPreset] = useState<OIDCPreset>('github_actions')
  const [repoFullName, setRepoFullName] = useState(initialRepoFullName ?? '')

  const [name, setName] = useState(
    initialRepoFullName
      ? defaultRepoPolicyName(initialRepoFullName, reservedNames)
      : ''
  )
  const [issuerUrl, setIssuerUrl] = useState(GITHUB_ACTIONS_ISSUER)
  const [audience, setAudience] = useState(githubAudience)
  const [role, setRole] = useState('org_read_only')
  const [tokenDurationSeconds, setTokenDurationSeconds] = useState('900')
  const [claimConditions, setClaimConditions] = useState<ClaimCondition[]>([
    {
      key: 'sub',
      value:
        initialRepoFullName && initialRepoDefaultBranch
          ? githubSubClaim(initialRepoFullName, initialRepoDefaultBranch)
          : '',
    },
  ])

  const [isNameDirty, setIsNameDirty] = useState(false)
  const [isSubDirty, setIsSubDirty] = useState(false)

  const isGithub = preset === 'github_actions'

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

  const setSubCondition = (value: string) =>
    setClaimConditions((prev) => {
      const hasSub = prev.some((condition) => condition.key.trim() === 'sub')
      return hasSub
        ? prev.map((condition) =>
            condition.key.trim() === 'sub' ? { ...condition, value } : condition
          )
        : [{ key: 'sub', value }, ...prev]
    })

  const selectPreset = (nextPreset: OIDCPreset) => {
    setPreset(nextPreset)
    if (nextPreset === 'custom') {
      setRepoFullName('')
      setIssuerUrl('')
      setAudience('')
      setRole('org_read_only')
      setTokenDurationSeconds('')
      setClaimConditions([{ key: 'sub', value: '' }])
      if (!isNameDirty) setName('')
      return
    }
    setIssuerUrl(GITHUB_ACTIONS_ISSUER)
    setAudience(githubAudience)
    setRole('org_read_only')
    setTokenDurationSeconds('900')
  }

  const selectRepo = (nextRepoFullName: string) => {
    setRepoFullName(nextRepoFullName)
    const branch = repos.find(
      (repo) => repo.full_name === nextRepoFullName
    )?.default_branch
    if (!isNameDirty) {
      setName(defaultRepoPolicyName(nextRepoFullName, reservedNames))
    }
    if (!isSubDirty && branch) {
      setSubCondition(githubSubClaim(nextRepoFullName, branch))
    }
  }

  const updateClaimCondition = (
    index: number,
    field: 'key' | 'value',
    value: string
  ) => {
    if (field === 'value' && claimConditions[index]?.key.trim() === 'sub') {
      setIsSubDirty(true)
    }
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

        {lockPreset ? null : (
          <Select
            labelProps={{ labelText: 'Provider' }}
            options={PRESET_OPTIONS}
            value={preset}
            onChange={(value) => selectPreset(value as OIDCPreset)}
          />
        )}

        {isGithub ? (
          hasVCSConnections === false && !isLoadingRepos ? (
            <Banner theme="warn">
              Connect a GitHub organization to fill this in from one of your
              repositories.{' '}
              <Link href={vcsConnectionsHref}>Manage VCS connections</Link>
            </Banner>
          ) : (
            <Select
              labelProps={{ labelText: 'Repository' }}
              options={repos.map((repo) => ({
                value: repo.full_name,
                label: repo.full_name,
                badge: { label: repo.default_branch },
              }))}
              value={repoFullName}
              onChange={(value) => selectRepo(value)}
              disabled={!!initialRepoFullName || isLoadingRepos}
              searchable
              placeholder={
                isLoadingRepos
                  ? 'Loading repositories...'
                  : 'Select a repository'
              }
            />
          )
        ) : null}

        <div className="flex flex-col gap-2">
          <Label htmlFor="policy-name">Name</Label>
          <Input
            id="policy-name"
            placeholder="ci-deploy"
            value={name}
            onChange={(e) => {
              setIsNameDirty(true)
              setName(e.target.value)
            }}
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
            placeholder="https://oidc.example.com"
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
                  placeholder="acme/app:main"
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
