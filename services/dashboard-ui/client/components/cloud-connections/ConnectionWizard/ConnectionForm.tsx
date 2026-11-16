import { useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
import { FormSelect } from '@/components/common/form/FormSelect'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { AWS_REGIONS } from '@/configs/cloud-regions'
import type {
  TAPIError,
  TCloudConnectionSummary,
  TCreateCloudConnectionRequest,
} from '@/types'
import { PrincipalSharing } from '../PrincipalSharing'
import { samePrincipal, sharedPrincipalConnections } from '../queries'
import { isImagePreset } from '../repository-scope'
import { connectionSchema, type TConnectionValues } from './schema'

const principalPrefixes: Record<TConnectionValues['preset'], string> = {
  stacks: 'nuon-stacks',
  ecr: 'nuon-ecr-pull',
  gar: 'nuon-gar-pull',
  acr: 'nuon-acr',
  custom: 'nuon-custom',
}

const defaultPrincipalName = (
  preset: TConnectionValues['preset'],
  suffix: string
) => `${principalPrefixes[preset]}-${suffix}`

const isDefaultPrincipalName = (name: string, suffix: string) =>
  Object.values(principalPrefixes).some(
    (prefix) => name === `${prefix}-${suffix}`
  )

const principalSuffix = () => {
  const alphabet = 'abcdefghijklmnopqrstuvwxyz0123456789'
  const bytes = crypto.getRandomValues(new Uint8Array(4))
  return Array.from(bytes, (byte) => alphabet[byte % alphabet.length]).join('')
}

const presetForPlatform = (platform?: 'aws' | 'gcp' | 'azure') =>
  platform === 'gcp' ? 'gar' : platform === 'azure' ? 'acr' : 'ecr'

const principalModeDescriptions: Record<
  'aws' | 'gcp' | 'azure',
  { new: string; existing: string }
> = {
  aws: {
    new: 'Setup creates a dedicated IAM role for this connection.',
    existing: 'Reuse a role another connection already uses.',
  },
  gcp: {
    new: 'Setup creates a dedicated service account for this connection.',
    existing: 'Reuse a service account another connection already uses.',
  },
  azure: {
    new: 'Setup creates a dedicated application for this connection.',
    existing: 'Reuse an application another connection already uses.',
  },
}

export const ConnectionForm = ({
  values,
  onCreate,
  isPending = false,
  error,
  imageRepository = false,
  connections = [],
  orgId,
}: {
  values?: Partial<TConnectionValues>
  onCreate: (body: TCreateCloudConnectionRequest) => void
  isPending?: boolean
  error?: TAPIError | null
  imageRepository?: boolean
  connections?: TCloudConnectionSummary[]
  orgId?: string
}) => {
  const [suffix] = useState(principalSuffix)
  const initialPreset = values?.preset ?? presetForPlatform(values?.platform)
  const defaultValues: TConnectionValues = {
    name: '',
    platform: 'aws',
    target_id: '',
    role_name: defaultPrincipalName(initialPreset, suffix),
    gcp_project_number: '',
    service_account_name: defaultPrincipalName(initialPreset, suffix),
    azure_tenant_id: '',
    principal: '',
    principal_choice: '',
    principal_mode: 'new',
    preset: presetForPlatform(values?.platform),
    ecr_region: '',
    gar_location: '',
    azure_resource_group: '',
    azure_registry_name: '',
    azure_registry_login_server: '',
    ...values,
  }
  const form = useForm({
    defaultValues,
    validators: { onMount: connectionSchema, onChange: connectionSchema },
    onSubmit: ({ value }) => {
      if (isPending) return
      const v = connectionSchema.parse(value)
      const platform = v.platform || 'aws'
      onCreate({
        name: v.name,
        platform,
        target_id: v.target_id,
        preset: v.preset,
        principal:
          platform === 'azure'
            ? v.principal!
            : v.principal_mode === 'existing' && v.principal
              ? v.principal
              : platform === 'gcp'
                ? `${v.service_account_name}@${v.target_id}.iam.gserviceaccount.com`
                : `arn:aws:iam::${v.target_id}:role/${v.role_name}`,
        ...(platform === 'gcp'
          ? {
              gcp_project_number: v.gcp_project_number,
              default_region: v.gar_location,
            }
          : {}),
        ...(platform === 'azure'
          ? {
              azure_tenant_id: v.azure_tenant_id,
              azure_registry_id: `/subscriptions/${v.target_id}/resourceGroups/${v.azure_resource_group}/providers/Microsoft.ContainerRegistry/registries/${v.azure_registry_name}`,
              azure_registry_login_server: v.azure_registry_login_server,
            }
          : {}),
        ...(v.preset === 'ecr' ? { default_region: v.ecr_region } : {}),
      })
    },
  })
  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const platform = useStore(form.store, (s) => s.values.platform)
  const preset = useStore(form.store, (s) => s.values.preset)
  const principalMode = useStore(form.store, (s) => s.values.principal_mode)
  const azureTenantId = useStore(form.store, (s) => s.values.azure_tenant_id)
  const principals = connections.filter(
    (connection, index, all) =>
      connection.platform === platform &&
      index === all.findIndex((other) => samePrincipal(other, connection))
  )
  const principalChoice = useStore(form.store, (s) => s.values.principal_choice)
  const selectedPrincipal = principals.find((c) => c.id === principalChoice)
  const identity = useStore(form.store, (s) =>
    s.values.platform === 'azure'
      ? s.values.principal || ''
      : s.values.platform === 'gcp'
        ? `${s.values.service_account_name}@${s.values.target_id}.iam.gserviceaccount.com`
        : `arn:aws:iam::${s.values.target_id}:role/${s.values.role_name}`
  )
  const shared = sharedPrincipalConnections(connections, {
    platform: platform || 'aws',
    principal: identity,
    azure_tenant_id: azureTenantId,
  })
  const isGCP = platform === 'gcp'
  const isAzure = platform === 'azure'
  const isImage = isImagePreset(preset)
  const identityFields: [keyof TConnectionValues, string, string, string?][] = [
    [
      'target_id',
      isAzure ? 'Subscription ID' : isGCP ? 'Project ID' : 'AWS account ID',
      isAzure
        ? '11111111-1111-4111-8111-111111111111'
        : isGCP
          ? 'acme-project'
          : '123456789012',
    ],
    ...(isAzure
      ? ([
          [
            'azure_tenant_id',
            'Microsoft Entra tenant ID',
            '22222222-2222-4222-8222-222222222222',
          ],
          [
            'principal',
            'Application (client) ID',
            '33333333-3333-4333-8333-333333333333',
            'Use an existing application and service principal. No client secret is needed.',
          ],
        ] as [keyof TConnectionValues, string, string, string?][])
      : isGCP
        ? ([
            [
              'gcp_project_number',
              'Project number',
              '123456789012',
              'The numeric project number, not the project ID.',
            ],
            [
              'service_account_name',
              'Service account name',
              defaultPrincipalName('gar', suffix),
              'Setup creates this service account, or updates its access if it already exists in the project.',
            ],
          ] as [keyof TConnectionValues, string, string, string?][])
        : ([
            [
              'role_name',
              'IAM role name',
              defaultPrincipalName(preset, suffix),
              'Setup creates this role, or updates its trust if it already exists in AWS.',
            ],
          ] as [keyof TConnectionValues, string, string, string?][])),
  ]
  const targetFields: [keyof TConnectionValues, string, string, string?][] =
    isAzure
      ? [
          ['azure_resource_group', 'Resource group', 'acme-images'],
          [
            'azure_registry_name',
            'Registry name',
            'acmeimages',
            'The registry resource name, not a repository name.',
          ],
          [
            'azure_registry_login_server',
            'Registry login server',
            'acmeimages.azurecr.io',
            'Copy the actual login server, including any DNS hash; omit https:// and repository paths.',
          ],
        ]
      : isGCP
        ? [
            [
              'gar_location',
              'GAR location',
              'us-central1',
              'The region or multi-region containing the repositories.',
            ],
          ]
        : []
  return (
    <form
      className="flex max-w-2xl flex-col gap-6"
      onSubmit={(event) => {
        event.preventDefault()
        form.handleSubmit()
      }}
    >
      {imageRepository && (
        <SectionHeader
          title="New pull identity"
          description="Create a connection for pulling images. No long-lived credentials are stored."
        />
      )}
      <FormErrorBanner error={error} fallback="Connection creation failed" />
      {!imageRepository && (
        <form.Field
          name="platform"
          listeners={{
            onChange: ({ value }) => {
              form.setFieldValue(
                'preset',
                value === 'gcp' ? 'gar' : value === 'azure' ? 'acr' : 'ecr'
              )
              form.setFieldValue('principal_mode', 'new')
              form.setFieldValue('principal_choice', '')
              form.setFieldValue('target_id', '')
              form.setFieldMeta('target_id', (meta) => ({
                ...meta,
                isTouched: false,
              }))
            },
          }}
        >
          {(field) => (
            <FormSelect
              field={field}
              id="connection-platform"
              labelProps={{ labelText: 'Cloud platform' }}
              disabled={isPending}
              options={[
                { value: 'aws', label: 'AWS' },
                { value: 'gcp', label: 'Google Cloud' },
                { value: 'azure', label: 'Azure' },
              ]}
            />
          )}
        </form.Field>
      )}
      <form.Field name="name">
        {(field) => (
          <FormInput
            field={field}
            id="connection-name"
            labelProps={{ labelText: 'Connection name' }}
            placeholder="acme-images"
            className="!font-mono"
            disabled={isPending}
          />
        )}
      </form.Field>
      {!imageRepository && !isGCP && !isAzure && (
        <form.Field
          name="preset"
          listeners={{
            onChange: ({ value }) => {
              const roleName = form.getFieldValue('role_name')
              if (
                typeof roleName === 'string' &&
                isDefaultPrincipalName(roleName, suffix)
              )
                form.setFieldValue(
                  'role_name',
                  defaultPrincipalName(value, suffix)
                )
            },
          }}
        >
          {(field) => (
            <FormSelect
              field={field}
              id="connection-access"
              labelProps={{ labelText: 'Access type' }}
              disabled={isPending}
              options={[
                { value: 'ecr', label: 'Image repositories' },
                { value: 'stacks', label: 'Manage install stacks' },
                { value: 'custom', label: 'Custom policy' },
              ]}
            />
          )}
        </form.Field>
      )}
      <section
        className="flex flex-col gap-4 border-t pt-6"
        aria-label="Principal"
      >
        <SectionHeader
          title="Principal"
          description={`The principal Nuon will use in ${isAzure ? 'Azure' : isGCP ? 'Google Cloud' : 'AWS'}.`}
        />
        {principals.length > 0 && (
          <form.Field
            name="principal_mode"
            listeners={{
              onChange: () => {
                form.setFieldValue('principal_choice', '')
                for (const [name] of identityFields) {
                  form.setFieldMeta(name, (meta) => ({
                    ...meta,
                    isTouched: false,
                  }))
                }
              },
            }}
          >
            {(field) => {
              const descriptions = principalModeDescriptions[platform || 'aws']
              return (
                <FormRadioGroup
                  field={field}
                  disabled={isPending}
                  options={[
                    {
                      value: 'new',
                      label: 'Create a new principal',
                      description: descriptions.new,
                    },
                    {
                      value: 'existing',
                      label: 'Use an existing principal',
                      description: descriptions.existing,
                    },
                  ]}
                />
              )
            }}
          </form.Field>
        )}
        {principalMode === 'existing' && (
          <form.Field
            name="principal_choice"
            listeners={{
              onChange: ({ value }) => {
                const selected = principals.find((c) => c.id === value)
                const preset = form.getFieldValue('preset') || 'ecr'
                form.setFieldValue(
                  'target_id',
                  selected?.target_id ||
                    (platform === defaultValues.platform
                      ? defaultValues.target_id
                      : '')
                )
                form.setFieldValue(
                  'role_name',
                  selected
                    ? selected.principal.split(':role/')[1]
                    : defaultPrincipalName(preset, suffix)
                )
                form.setFieldValue(
                  'service_account_name',
                  selected
                    ? selected.principal.split('@')[0]
                    : defaultPrincipalName(preset, suffix)
                )
                form.setFieldValue(
                  'gcp_project_number',
                  selected?.gcp_project_number ||
                    defaultValues.gcp_project_number
                )
                form.setFieldValue(
                  'azure_tenant_id',
                  selected?.azure_tenant_id || defaultValues.azure_tenant_id
                )
                form.setFieldValue(
                  'principal',
                  selected?.principal || defaultValues.principal
                )
                if (selected && isAzure) {
                  const registry = selected.azure_registry_id?.match(
                    /\/resourceGroups\/([^/]+)\/providers\/Microsoft.ContainerRegistry\/registries\/([^/]+)$/i
                  )
                  form.setFieldValue(
                    'azure_resource_group',
                    registry?.[1] || ''
                  )
                  form.setFieldValue('azure_registry_name', registry?.[2] || '')
                  form.setFieldValue(
                    'azure_registry_login_server',
                    selected.azure_registry_login_server || ''
                  )
                }
                for (const [name] of identityFields) {
                  form.setFieldMeta(name, (meta) => ({
                    ...meta,
                    isTouched: false,
                  }))
                }
              },
            }}
          >
            {(field) => (
              <FormSelect
                field={field}
                id="connection-existing-principal"
                labelProps={{ labelText: 'Existing principal' }}
                placeholder="Choose an existing principal"
                disabled={isPending}
                searchable
                options={principals.map((connection) => ({
                  value: connection.id,
                  label: connection.principal,
                  content: (
                    <span className="flex min-w-0 flex-col">
                      <Text family="mono" className="break-all">
                        {connection.principal}
                      </Text>
                      {platform === 'azure' && (
                        <Text variant="subtext" theme="neutral" family="mono">
                          Tenant {connection.azure_tenant_id}
                        </Text>
                      )}
                    </span>
                  ),
                  description: sharedPrincipalConnections(connections, {
                    platform: connection.platform,
                    principal: connection.principal,
                    azure_tenant_id: connection.azure_tenant_id,
                  })
                    .map((c) => c.name)
                    .join(', '),
                }))}
              />
            )}
          </form.Field>
        )}
        {selectedPrincipal && (
          <LabeledValue
            label={
              isAzure
                ? 'Subscription ID'
                : isGCP
                  ? 'Project ID'
                  : 'AWS account ID'
            }
          >
            <Text family="mono">{selectedPrincipal.target_id}</Text>
          </LabeledValue>
        )}
        {principalMode !== 'existing' &&
          identityFields.map(([name, label, placeholder, helperText]) => (
            <form.Field key={name} name={name}>
              {(field) => (
                <FormInput
                  field={field}
                  id={`connection-${name}`}
                  labelProps={{ labelText: label }}
                  placeholder={placeholder}
                  helperText={helperText}
                  className="!font-mono"
                  disabled={isPending}
                  spellCheck={false}
                  autoCapitalize="none"
                />
              )}
            </form.Field>
          ))}
        <PrincipalSharing
          connections={shared}
          orgId={orgId}
          sharedPermissions={!!selectedPrincipal}
        />
      </section>
      {isImage && (
        <section
          className="flex flex-col gap-4 border-t pt-6"
          aria-label="Registry target"
        >
          <SectionHeader
            title="Registry target"
            description="The registry coordinates this principal can be assigned to."
          />
          {preset === 'ecr' && (
            <form.Field name="ecr_region">
              {(field) => (
                <FormSelect
                  field={field}
                  id="connection-ecr-region"
                  labelProps={{ labelText: 'ECR region' }}
                  options={AWS_REGIONS.filter(
                    (region) => !/^(us-gov-|cn-)/.test(region.value)
                  ).map((region) => ({
                    value: region.value,
                    label: `${region.text} [${region.value}]`,
                    content: (
                      <>
                        {region.text}{' '}
                        <Text family="mono">[{region.value}]</Text>
                      </>
                    ),
                  }))}
                  placeholder="Choose ECR region"
                  helperText="GovCloud and China are not supported."
                  searchable
                  disabled={isPending}
                />
              )}
            </form.Field>
          )}
          {targetFields.map(([name, label, placeholder, helperText]) => (
            <form.Field key={name} name={name}>
              {(field) => (
                <FormInput
                  field={field}
                  id={`connection-${name}`}
                  labelProps={{ labelText: label }}
                  placeholder={placeholder}
                  helperText={helperText}
                  className="!font-mono"
                  disabled={isPending}
                  spellCheck={false}
                  autoCapitalize="none"
                />
              )}
            </form.Field>
          ))}
        </section>
      )}
      <div className="flex flex-wrap items-center justify-between gap-4 border-t pt-4">
        <Text variant="subtext" theme="neutral">
          {imageRepository
            ? 'Creating a pull identity does not change cloud permissions.'
            : 'Creating a connection does not change cloud permissions.'}
        </Text>
        <Button
          type="submit"
          variant="primary"
          disabled={!canSubmit || isPending}
          tooltipProps={
            !canSubmit && !isPending
              ? {
                  tipContent: `Cannot create ${imageRepository ? 'pull identity' : 'connection'} — complete the required details`,
                }
              : undefined
          }
        >
          {isPending ? (
            <>
              <Icon variant="Loading" />{' '}
              {imageRepository
                ? 'Creating pull identity'
                : 'Creating connection'}
            </>
          ) : imageRepository ? (
            'Create pull identity'
          ) : (
            'Create connection'
          )}
        </Button>
      </div>
    </form>
  )
}
