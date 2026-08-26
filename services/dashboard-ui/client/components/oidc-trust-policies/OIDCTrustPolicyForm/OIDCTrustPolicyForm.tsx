import { useMemo, useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import type { FormValidateOrFn } from '@tanstack/form-core'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormSelect } from '@/components/common/form/FormSelect'
import { FormToggle } from '@/components/common/form/FormToggle'
import { Select } from '@/components/common/form/Select'
import { fieldErrorMessage } from '@/components/common/form/field-error'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TOIDCTrustPolicy, TVCSConnectionRepo } from '@/types'
import {
  buildOIDCSchema,
  defaultRepoPolicyName,
  GITHUB_ACTIONS_ISSUER,
  githubSubClaim,
  hasSubCondition,
  type ClaimCondition,
  type OIDCFormValues,
  type OIDCPreset,
  type OIDCTrustPolicyMode,
} from './schema'

const PRESET_OPTIONS = [
  { value: 'github_actions', label: 'GitHub Actions' },
  { value: 'custom', label: 'Custom' },
]

const conditionsToRows = (
  claimConditions: TOIDCTrustPolicy['claim_conditions']
): ClaimCondition[] => {
  const entries = Object.entries(claimConditions ?? {})
  return entries.length
    ? entries.map(([key, value]) => ({ key, value }))
    : [{ key: 'sub', value: '' }]
}

const buildDefaultValues = ({
  mode,
  policy,
  initialRepoFullName,
  initialRepoDefaultBranch,
  githubAudience,
  reservedNames,
  defaultRole,
  defaultName,
}: {
  mode: OIDCTrustPolicyMode
  policy?: TOIDCTrustPolicy
  initialRepoFullName?: string
  initialRepoDefaultBranch?: string
  githubAudience: string
  reservedNames?: string[]
  defaultRole?: string
  defaultName?: string
}): OIDCFormValues => {
  if (mode === 'edit' && policy) {
    return {
      name: policy.name ?? '',
      issuerUrl: policy.issuer_url ?? '',
      audience: policy.audience ?? '',
      role: policy.role ?? 'org_read_only',
      tokenDurationSeconds: policy.token_duration_seconds
        ? String(policy.token_duration_seconds)
        : '',
      enabled: policy.enabled ?? true,
      claimConditions: conditionsToRows(policy.claim_conditions),
    }
  }
  return {
    name:
      defaultName ??
      (initialRepoFullName
        ? defaultRepoPolicyName(initialRepoFullName, reservedNames)
        : ''),
    issuerUrl: GITHUB_ACTIONS_ISSUER,
    audience: githubAudience,
    role: defaultRole ?? 'org_read_only',
    tokenDurationSeconds: '900',
    enabled: true,
    claimConditions: [
      {
        key: 'sub',
        value:
          initialRepoFullName && initialRepoDefaultBranch
            ? githubSubClaim(initialRepoFullName, initialRepoDefaultBranch)
            : '',
      },
    ],
  }
}

export const OIDCTrustPolicyFormModal = ({
  mode,
  policy,
  isPending,
  error,
  onSubmit,
  roleOptions,
  repos = [],
  isLoadingRepos,
  hasVCSConnections,
  vcsConnectionsHref = '',
  githubAudience = '',
  initialRepoFullName,
  initialRepoDefaultBranch,
  lockPreset,
  reservedNames,
  repoSource = 'connections',
  defaultRole,
  defaultName,
  ...props
}: {
  mode: OIDCTrustPolicyMode
  policy?: TOIDCTrustPolicy
  isPending: boolean
  error: TAPIError | null
  onSubmit: (input: OIDCFormValues) => void
  roleOptions: { value: string; label: string; description?: string }[]
  repos?: TVCSConnectionRepo[]
  isLoadingRepos?: boolean
  hasVCSConnections?: boolean
  vcsConnectionsHref?: string
  githubAudience?: string
  initialRepoFullName?: string
  initialRepoDefaultBranch?: string
  lockPreset?: boolean
  reservedNames?: string[]
  // 'manual' takes a typed owner/repo, for policies about a repository this org has
  // no connection to — a customer's, in the install-stack flow.
  repoSource?: 'connections' | 'manual'
  defaultRole?: string
  defaultName?: string
} & Omit<IModal, 'onSubmit'>) => {
  const [preset, setPreset] = useState<OIDCPreset>('github_actions')
  const [repoFullName, setRepoFullName] = useState(initialRepoFullName ?? '')
  const [isNameDirty, setIsNameDirty] = useState(false)
  const [isSubDirty, setIsSubDirty] = useState(false)
  const [manualBranch, setManualBranch] = useState(
    initialRepoDefaultBranch ?? 'main'
  )

  const schema = useMemo(
    () => buildOIDCSchema({ mode, reservedNames }),
    [mode, reservedNames]
  )
  const validator = schema as unknown as FormValidateOrFn<OIDCFormValues>

  const form = useForm({
    defaultValues: buildDefaultValues({
      mode,
      policy,
      initialRepoFullName,
      initialRepoDefaultBranch,
      githubAudience,
      reservedNames,
      defaultRole,
      defaultName,
    }),
    validators: { onMount: validator, onChange: validator },
    onSubmit: ({ value }) =>
      onSubmit({
        name: value.name.trim(),
        issuerUrl: value.issuerUrl.trim(),
        audience: value.audience.trim(),
        role: value.role,
        tokenDurationSeconds: value.tokenDurationSeconds,
        enabled: value.enabled,
        claimConditions: value.claimConditions,
      }),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const claimConditions = useStore(form.store, (s) => s.values.claimConditions)
  const hasSub = hasSubCondition(claimConditions)

  const isCreate = mode === 'create'
  const isGithub = preset === 'github_actions'

  const setSubCondition = (value: string) => {
    const prev = form.getFieldValue('claimConditions') as ClaimCondition[]
    const existing = prev.some((condition) => condition.key.trim() === 'sub')
    const next = existing
      ? prev.map((condition) =>
          condition.key.trim() === 'sub' ? { ...condition, value } : condition
        )
      : [{ key: 'sub', value }, ...prev]
    form.setFieldValue('claimConditions', next)
  }

  const selectPreset = (nextPreset: OIDCPreset) => {
    setPreset(nextPreset)
    if (nextPreset === 'custom') {
      setRepoFullName('')
      form.setFieldValue('issuerUrl', '')
      form.setFieldValue('audience', '')
      form.setFieldValue('role', defaultRole ?? 'org_read_only')
      form.setFieldValue('tokenDurationSeconds', '')
      form.setFieldValue('claimConditions', [{ key: 'sub', value: '' }])
      if (!isNameDirty) form.setFieldValue('name', '')
      return
    }
    form.setFieldValue('issuerUrl', GITHUB_ACTIONS_ISSUER)
    form.setFieldValue('audience', githubAudience)
    form.setFieldValue('role', defaultRole ?? 'org_read_only')
    form.setFieldValue('tokenDurationSeconds', '900')
  }

  // Shared by both repo modes: a picked repo and a typed one derive the policy name
  // and the sub claim identically.
  const applyRepo = (nextRepoFullName: string, branch?: string) => {
    setRepoFullName(nextRepoFullName)
    if (!isNameDirty && !defaultName) {
      form.setFieldValue(
        'name',
        defaultRepoPolicyName(nextRepoFullName, reservedNames)
      )
    }
    if (!isSubDirty && nextRepoFullName && branch) {
      setSubCondition(githubSubClaim(nextRepoFullName, branch))
    }
  }

  const selectRepo = (nextRepoFullName: string) =>
    applyRepo(
      nextRepoFullName,
      repos.find((repo) => repo.full_name === nextRepoFullName)?.default_branch
    )

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="ShieldCheckIcon" size="24" />
          {isCreate ? 'Create OIDC trust policy' : 'Edit trust policy'}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            {isCreate ? 'Creating trust policy' : 'Saving changes'}
          </span>
        ) : isCreate ? (
          <span className="flex items-center gap-2">
            <Icon variant="PlusIcon" />
            Create trust policy
          </span>
        ) : (
          'Save changes'
        ),
        disabled: !canSubmit || !hasSub || isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-6"
      >
        <FormErrorBanner
          error={error}
          fallback={
            isCreate
              ? 'Unable to create trust policy'
              : 'Unable to update trust policy'
          }
        />

        {!isCreate && (
          <form.Field name="enabled">
            {(field) => (
              <FormToggle
                field={field}
                label="Enabled"
                description="Disabled policies reject token exchange requests."
              />
            )}
          </form.Field>
        )}

        {isCreate && !lockPreset && (
          <Select
            labelProps={{ labelText: 'Provider' }}
            options={PRESET_OPTIONS}
            value={preset}
            onChange={(value) => selectPreset(value as OIDCPreset)}
          />
        )}

        {isCreate && isGithub ? (
          repoSource === 'manual' ? (
            // No repo Select: the repository belongs to whoever runs the workflow, not this org.
            <div className="flex flex-col gap-2">
              <Label htmlFor="policy-repo">Repository</Label>
              <Input
                id="policy-repo"
                placeholder="acme/infra"
                value={repoFullName}
                onChange={(e) => applyRepo(e.target.value, manualBranch)}
                disabled={isPending}
              />
              <Label htmlFor="policy-branch">Branch</Label>
              <Input
                id="policy-branch"
                placeholder="main"
                value={manualBranch}
                onChange={(e) => {
                  setManualBranch(e.target.value)
                  applyRepo(repoFullName, e.target.value)
                }}
                disabled={isPending}
              />
              <Text variant="subtext" theme="neutral">
                Sets the <code>sub</code> claim below. Edit that directly for
                anything other than a branch — a tag, an environment, or a
                wildcard across branches.
              </Text>
            </div>
          ) : hasVCSConnections === false && !isLoadingRepos ? (
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

        {isCreate ? (
          <form.Field name="name">
            {(field) => {
              const message = fieldErrorMessage(field)
              return (
                <div className="flex flex-col gap-2">
                  <Label htmlFor="policy-name">Name</Label>
                  <Input
                    id="policy-name"
                    placeholder="ci-deploy"
                    value={field.state.value}
                    onChange={(e) => {
                      setIsNameDirty(true)
                      field.handleChange(e.target.value)
                    }}
                    onBlur={field.handleBlur}
                    disabled={isPending}
                    error={!!message}
                    errorMessage={message}
                  />
                </div>
              )
            }}
          </form.Field>
        ) : (
          <form.Field name="name">
            {(field) => (
              <FormInput
                field={field}
                id="policy-name"
                disabled={isPending}
                labelProps={{ labelText: 'Name' }}
              />
            )}
          </form.Field>
        )}

        <form.Field name="issuerUrl">
          {(field) => (
            <FormInput
              field={field}
              id="policy-issuer-url"
              placeholder="https://oidc.example.com"
              type="url"
              disabled={isPending}
              labelProps={{ labelText: 'Issuer URL' }}
              helperText="Must be an absolute http or https URL."
            />
          )}
        </form.Field>

        <form.Field name="audience">
          {(field) => (
            <FormInput
              field={field}
              id="policy-audience"
              placeholder="https://api.nuon.co"
              disabled={isPending}
              labelProps={{ labelText: 'Audience' }}
              helperText="The expected `aud` claim value on the presented token."
            />
          )}
        </form.Field>

        <form.Field name="role">
          {(field) => (
            <FormSelect
              field={field}
              options={roleOptions}
              disabled={isPending}
              labelProps={{ labelText: 'Role' }}
              helperText="Org role granted to tokens exchanged with this policy."
            />
          )}
        </form.Field>

        <form.Field name="tokenDurationSeconds">
          {(field) => (
            <FormInput
              field={field}
              id="policy-token-duration"
              placeholder="3600"
              type="number"
              min={1}
              max={86400}
              disabled={isPending}
              labelProps={{ labelText: 'Token duration in seconds (optional)' }}
              helperText="Maximum is 86400."
            />
          )}
        </form.Field>

        <div className="flex flex-col gap-2">
          <Label>Claim conditions</Label>
          <Text variant="subtext" theme="neutral">
            All conditions must match the presented token. A `sub` condition is
            required.
          </Text>
          <form.Field name="claimConditions" mode="array">
            {(ccField) => (
              <>
                <div className="flex flex-col gap-2">
                  {ccField.state.value.map((row, index) => (
                    <div key={index} className="flex items-center gap-2">
                      <form.Field name={`claimConditions[${index}].key`}>
                        {(f) => (
                          <Input
                            placeholder="sub"
                            value={f.state.value}
                            onChange={(e) => f.handleChange(e.target.value)}
                            onBlur={f.handleBlur}
                            disabled={isPending}
                          />
                        )}
                      </form.Field>
                      <form.Field name={`claimConditions[${index}].value`}>
                        {(f) => (
                          <Input
                            placeholder="repo:acme/app:ref:refs/heads/main"
                            value={f.state.value}
                            onChange={(e) => {
                              if (isCreate && row.key.trim() === 'sub')
                                setIsSubDirty(true)
                              f.handleChange(e.target.value)
                            }}
                            onBlur={f.handleBlur}
                            disabled={isPending}
                          />
                        )}
                      </form.Field>
                      <Button
                        variant="icon"
                        aria-label="Remove claim condition"
                        disabled={ccField.state.value.length === 1 || isPending}
                        onClick={() => ccField.removeValue(index)}
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
                  disabled={isPending}
                  onClick={() => ccField.pushValue({ key: '', value: '' })}
                >
                  <Icon variant="PlusIcon" size={14} />
                  Add condition
                </Button>
              </>
            )}
          </form.Field>
        </div>
      </form>
    </Modal>
  )
}
