import { useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormSelect } from '@/components/common/form/FormSelect'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { ConnectionForm } from '@/components/cloud-connections/ConnectionWizard/ConnectionForm'
import {
  canonicalRepositories,
  repositoryNamesSchema,
} from '@/components/cloud-connections/repository-scope'
import type {
  TAPIError,
  TCloudConnectionSummary,
  TCreateCloudConnectionRequest,
  TCreateImageRepositoryRequest,
} from '@/types'

const schema = z
  .object({
    platform: z.enum(['aws', 'gcp', 'azure']),
    repository: z.string().trim().min(1, 'Enter a repository name'),
    connection_id: z.string().min(1, 'Choose an identity or create one'),
  })
  .superRefine((value, ctx) => {
    const preset =
      value.platform === 'aws'
        ? 'ecr'
        : value.platform === 'gcp'
          ? 'gar'
          : 'acr'
    const result = repositoryNamesSchema(preset).safeParse(value.repository)
    if (!result.success || /[\r\n]/.test(value.repository))
      ctx.addIssue({
        code: 'custom',
        path: ['repository'],
        message: result.success
          ? 'Enter one repository name'
          : result.error.issues[0].message,
      })
  })

export const AddImageRepository = ({
  orgId,
  connections,
  isLoading = false,
  isPending = false,
  isCreatingIdentity = false,
  error,
  identityError,
  onSave,
  onCreateIdentity,
  initialValues,
}: {
  orgId: string
  connections: TCloudConnectionSummary[]
  isLoading?: boolean
  isPending?: boolean
  isCreatingIdentity?: boolean
  error?: TAPIError | null
  identityError?: TAPIError | null
  onSave: (body: TCreateImageRepositoryRequest) => void
  onCreateIdentity: (
    body: TCreateCloudConnectionRequest
  ) => Promise<TCloudConnectionSummary>
  initialValues?: {
    platform: 'aws' | 'gcp' | 'azure'
    repository: string
    connection_id: string
  }
}) => {
  const [creatingIdentity, setCreatingIdentity] = useState(false)
  const [createdIdentity, setCreatedIdentity] =
    useState<TCloudConnectionSummary>()
  const form = useForm({
    defaultValues: initialValues || {
      platform: 'aws' as const,
      repository: '',
      connection_id: '',
    },
    validators: { onMount: schema, onChange: schema },
    onSubmit: ({ value }) => {
      const identity = [createdIdentity, ...connections].find(
        (c) => c?.id === value.connection_id
      )
      if (!identity || isPending) return
      onSave({
        cloud_connection_id: identity.id,
        repository: canonicalRepositories(identity, [
          value.repository.trim(),
        ])[0],
      })
    },
  })
  const platform = useStore(form.store, (state) => state.values.platform)
  const connectionId = useStore(
    form.store,
    (state) => state.values.connection_id
  )
  const repository = useStore(form.store, (state) => state.values.repository)
  const canSubmit = useStore(form.store, (state) => state.canSubmit)
  const preset = platform === 'aws' ? 'ecr' : platform === 'gcp' ? 'gar' : 'acr'
  const choices = [
    ...connections.filter((c) => c.id !== createdIdentity?.id),
    ...(createdIdentity ? [createdIdentity] : []),
  ].filter(
    (c) =>
      c.platform === platform &&
      c.preset === preset &&
      (preset === 'acr' || !!c.default_region)
  )
  const selected = choices.find((c) => c.id === connectionId)
  const imagePrefix =
    selected && repository.trim()
      ? platform === 'aws'
        ? `${selected.target_id}.dkr.ecr.${selected.default_region}.amazonaws.com/${repository.trim()}`
        : platform === 'gcp'
          ? `${selected.default_region}-docker.pkg.dev/${selected.target_id}/${repository.trim()}`
          : `${selected.azure_registry_login_server}/${repository.trim()}`
      : undefined
  return (
    <PageSection>
      <SectionHeader
        title="Add image repository"
        description="Save a private image source and choose the identity Nuon will use to pull it."
      />
      {creatingIdentity ? (
        <div className="flex max-w-2xl flex-col gap-6">
          <Button
            className="w-fit"
            disabled={isCreatingIdentity}
            onClick={() => setCreatingIdentity(false)}
          >
            Use an existing identity
          </Button>
          <Text theme="neutral">
            {platform === 'aws'
              ? 'Amazon ECR'
              : platform === 'gcp'
                ? 'Google Artifact Registry'
                : 'Azure Container Registry'}
          </Text>
          <ConnectionForm
            imageRepository
            orgId={orgId}
            connections={connections}
            values={{ platform, preset }}
            isPending={isCreatingIdentity}
            error={identityError}
            onCreate={async (body) => {
              try {
                const identity = await onCreateIdentity(body)
                setCreatedIdentity(identity)
                form.setFieldValue('connection_id', identity.id)
                setCreatingIdentity(false)
              } catch {
                return
              }
            }}
          />
        </div>
      ) : (
        <form
          className="flex max-w-2xl flex-col gap-6"
          onSubmit={(event) => {
            event.preventDefault()
            form.handleSubmit()
          }}
        >
          <FormErrorBanner
            error={error}
            fallback="Repository could not be saved"
          />
          <form.Field
            name="platform"
            listeners={{
              onChange: () => {
                form.setFieldValue('connection_id', '')
                form.setFieldValue('repository', '')
              },
            }}
          >
            {(field) => (
              <FormSelect
                field={field}
                id="image-repository-platform"
                labelProps={{ labelText: 'Registry provider' }}
                disabled={isPending}
                options={[
                  { value: 'aws', label: 'Amazon ECR' },
                  { value: 'gcp', label: 'Google Artifact Registry' },
                  { value: 'azure', label: 'Azure Container Registry' },
                ]}
              />
            )}
          </form.Field>
          <form.Field name="repository">
            {(field) => (
              <FormInput
                field={field}
                id="image-repository-name"
                labelProps={{
                  labelText:
                    platform === 'gcp'
                      ? 'GAR repository ID'
                      : 'Repository name',
                }}
                helperText={
                  platform === 'gcp'
                    ? 'Enter the Artifact Registry repository ID, not an image path or tag.'
                    : 'Include any nested path. Omit the registry hostname, tag, and digest.'
                }
                placeholder={platform === 'gcp' ? 'acme-services' : 'acme/api'}
                className="!font-mono"
                disabled={isPending}
                autoCapitalize="none"
                spellCheck={false}
              />
            )}
          </form.Field>
          <section
            className="flex flex-col gap-4 border-t pt-6"
            aria-label="Authentication"
          >
            <SectionHeader
              title="Authentication"
              description="Use an existing identity or create one for this repository. Identities are managed as cloud connections."
            />
            <form.Field name="connection_id">
              {(field) => (
                <FormSelect
                  field={field}
                  id="image-repository-identity"
                  labelProps={{ labelText: 'Pull identity' }}
                  placeholder={
                    isLoading ? 'Loading identities...' : 'Choose an identity'
                  }
                  disabled={isLoading || isPending}
                  searchable
                  options={choices.map((identity) => ({
                    value: identity.id,
                    label: identity.name,
                    description: [
                      identity.target_id,
                      identity.default_region,
                      identity.azure_registry_login_server,
                    ]
                      .filter(Boolean)
                      .join(' · '),
                    detail: <Status status={identity.status} variant="badge" />,
                    searchText: `${identity.name} ${identity.target_id} ${identity.principal} ${identity.default_region || ''} ${identity.azure_registry_login_server || ''}`,
                  }))}
                />
              )}
            </form.Field>
            {!isLoading && choices.length === 0 && (
              <Text variant="subtext" theme="neutral">
                No identities are available for this provider. Create one to
                continue.
              </Text>
            )}
            <Button
              className="w-fit"
              disabled={isPending}
              onClick={() => setCreatingIdentity(true)}
            >
              Create new identity
            </Button>
            {selected && (
              <LabeledValue
                label={
                  platform === 'aws'
                    ? 'IAM role'
                    : platform === 'gcp'
                      ? 'Service account'
                      : 'Application (client) ID'
                }
              >
                <Text family="mono" variant="subtext" className="break-all">
                  {selected.principal}
                </Text>
              </LabeledValue>
            )}
          </section>
          {imagePrefix && (
            <LabeledValue label="Repository location">
              <Text family="mono" className="break-all">
                {imagePrefix}
              </Text>
            </LabeledValue>
          )}
          <Banner theme="info">
            Nuon saves this repository and the selected identity. Saving does
            not grant cloud permissions; review and apply the setup commands
            next.
          </Banner>
          <div className="flex flex-wrap justify-end gap-2 border-t pt-4">
            <Button href={`/${orgId}/settings/image-repositories`}>
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              disabled={!canSubmit || !selected || isPending}
              tooltipProps={
                !canSubmit || !selected
                  ? {
                      tipContent:
                        'Enter a valid repository and choose a pull identity',
                    }
                  : undefined
              }
            >
              {isPending ? 'Saving repository' : 'Save repository'}
            </Button>
          </div>
        </form>
      )}
    </PageSection>
  )
}
