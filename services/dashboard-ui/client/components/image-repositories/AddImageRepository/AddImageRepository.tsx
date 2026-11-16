import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
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

const presetForPlatform = (platform: 'aws' | 'gcp' | 'azure') =>
  platform === 'aws' ? 'ecr' : platform === 'gcp' ? 'gar' : 'acr'

const schema = z
  .object({
    platform: z.enum(['aws', 'gcp', 'azure']),
    repository: z.string().trim().min(1, 'Repository name is required'),
    connection_id: z.string().min(1, 'Pull identity is required'),
    identity_mode: z.enum(['existing', 'new']),
  })
  .superRefine((value, ctx) => {
    const preset = presetForPlatform(value.platform)
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
    identity_mode?: 'existing' | 'new'
  }
}) => {
  const hasIdentities = (platform: 'aws' | 'gcp' | 'azure') => {
    const preset = presetForPlatform(platform)
    return connections.some(
      (c) =>
        c.platform === platform &&
        c.preset === preset &&
        (preset === 'acr' || !!c.default_region)
    )
  }
  const initialPlatform = initialValues?.platform ?? 'aws'
  const form = useForm({
    defaultValues: {
      platform: initialPlatform,
      repository: initialValues?.repository ?? '',
      connection_id: initialValues?.connection_id ?? '',
      identity_mode:
        initialValues?.identity_mode ??
        (hasIdentities(initialPlatform) ? 'existing' : 'new'),
    },
    validators: { onMount: schema, onChange: schema },
    onSubmit: ({ value }) => {
      const identity = connections.find((c) => c?.id === value.connection_id)
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
  const preset = presetForPlatform(platform)
  const identityMode = useStore(form.store, (s) => s.values.identity_mode)
  const choices = connections.filter(
    (c) =>
      c.platform === platform &&
      c.preset === preset &&
      (preset === 'acr' || !!c.default_region)
  )
  const showModeChoice = choices.length > 0
  const showCreate = identityMode === 'new' || (!isLoading && !showModeChoice)
  const modeRadio = showModeChoice && (
    <form.Field name="identity_mode">
      {(field) => (
        <FormRadioGroup
          field={field}
          disabled={isPending}
          options={[
            { value: 'new', label: 'Create a new principal' },
            { value: 'existing', label: 'Use an existing principal' },
          ]}
        />
      )}
    </form.Field>
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
        description="Save a private image source and choose the pull identity Nuon will use to fetch it."
      />
      <div className="flex max-w-2xl flex-col gap-6">
        <FormErrorBanner error={error} fallback="Repository save failed" />
        <form.Field
          name="platform"
          listeners={{
            onChange: ({ value }) => {
              form.setFieldValue('connection_id', '')
              form.setFieldValue('repository', '')
              form.setFieldValue(
                'identity_mode',
                hasIdentities(value) ? 'existing' : 'new'
              )
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
                  platform === 'gcp' ? 'GAR repository ID' : 'Repository name',
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
            description="Pull identities are saved as connections."
          />
          {modeRadio}
          {showCreate ? (
            <ConnectionForm
              key={platform}
              imageRepository
              orgId={orgId}
              connections={connections}
              values={{ platform, preset }}
              isPending={isCreatingIdentity}
              error={identityError}
              onCreate={async (body) => {
                try {
                  const identity = await onCreateIdentity(body)
                  form.setFieldValue('connection_id', identity.id)
                  form.setFieldValue('identity_mode', 'existing')
                } catch {
                  return
                }
              }}
            />
          ) : (
            <>
              <form.Field name="connection_id">
                {(field) => (
                  <FormSelect
                    field={field}
                    id="image-repository-identity"
                    labelProps={{ labelText: 'Pull identity' }}
                    placeholder={
                      isLoading
                        ? 'Loading pull identities...'
                        : 'Choose a pull identity'
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
                      detail: (
                        <Status status={identity.status} variant="badge" />
                      ),
                      searchText: `${identity.name} ${identity.target_id} ${identity.principal} ${identity.default_region || ''} ${identity.azure_registry_login_server || ''}`,
                    }))}
                  />
                )}
              </form.Field>
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
            </>
          )}
        </section>
        {!showCreate && imagePrefix && (
          <LabeledValue label="Repository location">
            <Text family="mono" className="break-all">
              {imagePrefix}
            </Text>
          </LabeledValue>
        )}
        {!showCreate && (
          <Banner theme="info">
            Nuon saves this repository and the selected pull identity. Saving
            does not grant cloud permissions; review and apply the setup
            commands next.
          </Banner>
        )}
        <div className="flex flex-wrap justify-end gap-2 border-t pt-4">
          <Button href={`/${orgId}/settings/image-repositories`}>Cancel</Button>
          {!showCreate && (
            <Button
              variant="primary"
              disabled={!canSubmit || !selected || isPending}
              tooltipProps={
                !canSubmit || !selected
                  ? {
                      tipContent:
                        'Cannot save — enter a repository name and choose a pull identity',
                    }
                  : undefined
              }
              onClick={() => form.handleSubmit()}
            >
              {isPending ? 'Saving repository' : 'Save repository'}
            </Button>
          )}
        </div>
      </div>
    </PageSection>
  )
}
