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
import { Select } from '@/components/common/form/Select'
import { fieldErrorMessage } from '@/components/common/form/field-error'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TVCSConnectionRepo } from '@/types'
import {
  buildCreateOIDCSchema,
  defaultRepoPolicyName,
  GITHUB_ACTIONS_ISSUER,
  githubSubClaim,
  hasSubCondition,
  type ClaimCondition,
  type OIDCFormValues,
  type OIDCPreset,
} from './schema'

export {
  hasSubCondition,
  type ClaimCondition,
  type OIDCTrustPolicyFormInput,
} from './schema'

const PRESET_OPTIONS = [
  { value: 'github_actions', label: 'GitHub Actions' },
  { value: 'custom', label: 'Custom' },
]

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
  onSubmit: (input: OIDCFormValues) => void
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
  const [isNameDirty, setIsNameDirty] = useState(false)
  const [isSubDirty, setIsSubDirty] = useState(false)

  const schema = useMemo(
    () => buildCreateOIDCSchema(reservedNames),
    [reservedNames]
  )
  const validator = schema as unknown as FormValidateOrFn<OIDCFormValues>

  const form = useForm({
    defaultValues: {
      name: initialRepoFullName
        ? defaultRepoPolicyName(initialRepoFullName, reservedNames)
        : '',
      issuerUrl: GITHUB_ACTIONS_ISSUER,
      audience: githubAudience,
      role: 'org_read_only',
      tokenDurationSeconds: '900',
      claimConditions: [
        {
          key: 'sub',
          value:
            initialRepoFullName && initialRepoDefaultBranch
              ? githubSubClaim(initialRepoFullName, initialRepoDefaultBranch)
              : '',
        },
      ],
    } as OIDCFormValues,
    validators: { onMount: validator, onChange: validator },
    onSubmit: ({ value }) =>
      onSubmit({
        name: value.name.trim(),
        issuerUrl: value.issuerUrl.trim(),
        audience: value.audience.trim(),
        role: value.role,
        tokenDurationSeconds: value.tokenDurationSeconds,
        claimConditions: value.claimConditions,
      }),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const claimConditions = useStore(form.store, (s) => s.values.claimConditions)
  const hasSub = hasSubCondition(claimConditions)

  const isGithub = preset === 'github_actions'

  const setSubCondition = (value: string) => {
    const prev = form.getFieldValue('claimConditions') as ClaimCondition[]
    const hasSub = prev.some((condition) => condition.key.trim() === 'sub')
    const next = hasSub
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
      form.setFieldValue('role', 'org_read_only')
      form.setFieldValue('tokenDurationSeconds', '')
      form.setFieldValue('claimConditions', [{ key: 'sub', value: '' }])
      if (!isNameDirty) form.setFieldValue('name', '')
      return
    }
    form.setFieldValue('issuerUrl', GITHUB_ACTIONS_ISSUER)
    form.setFieldValue('audience', githubAudience)
    form.setFieldValue('role', 'org_read_only')
    form.setFieldValue('tokenDurationSeconds', '900')
  }

  const selectRepo = (nextRepoFullName: string) => {
    setRepoFullName(nextRepoFullName)
    const branch = repos.find(
      (repo) => repo.full_name === nextRepoFullName
    )?.default_branch
    if (!isNameDirty) {
      form.setFieldValue(
        'name',
        defaultRepoPolicyName(nextRepoFullName, reservedNames)
      )
    }
    if (!isSubDirty && branch) {
      setSubCondition(githubSubClaim(nextRepoFullName, branch))
    }
  }

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
            <Icon variant="Loading" /> Creating trust policy
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="PlusIcon" />
            Create trust policy
          </span>
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
        <FormErrorBanner error={error} fallback="Unable to create trust policy" />

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
                            placeholder="acme/app:main"
                            value={f.state.value}
                            onChange={(e) => {
                              if (row.key.trim() === 'sub') setIsSubDirty(true)
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
