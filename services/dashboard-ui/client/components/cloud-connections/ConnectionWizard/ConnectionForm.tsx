import { useForm, useStore } from '@tanstack/react-form'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormSelect } from '@/components/common/form/FormSelect'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { AWS_REGIONS } from '@/configs/cloud-regions'
import type {
  TAPIError,
  TCloudConnectionSummary,
  TCreateCloudConnectionRequest,
} from '@/types'
import { PrincipalSharing } from '../PrincipalSharing'
import { sharedPrincipalConnections } from '../queries'
import { connectionSchema, type TConnectionValues } from './schema'

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
  const defaultValues: TConnectionValues = {
    name: '',
    platform: 'aws',
    target_id: '',
    role_name: 'nuon-cloud-connection',
    gcp_project_number: '',
    service_account_name: 'nuon-cloud-connection',
    azure_tenant_id: '',
    principal: '',
    principal_choice: values?.target_id ? 'new' : '',
    preset:
      values?.platform === 'gcp'
        ? 'gar'
        : values?.platform === 'azure'
          ? 'acr'
          : 'ecr',
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
          platform === 'azure' ||
          (v.principal_choice && v.principal_choice !== 'new')
            ? v.principal!
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
  const principalChoice = useStore(form.store, (s) => s.values.principal_choice)
  const azureTenantId = useStore(form.store, (s) => s.values.azure_tenant_id)
  const principals = connections.filter(
    (connection, index, all) =>
      connection.platform === platform &&
      index ===
        all.findIndex(
          (other) =>
            other.platform === connection.platform &&
            (connection.platform === 'azure'
              ? other.principal.toLowerCase() ===
                  connection.principal.toLowerCase() &&
                other.azure_tenant_id?.toLowerCase() ===
                  connection.azure_tenant_id?.toLowerCase()
              : other.principal === connection.principal)
        )
  )
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
  const isImage = ['ecr', 'gar', 'acr'].includes(preset)
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
              'nuon-cloud-connection',
              'Use 6–30 lowercase letters, numbers, or dashes, starting with a letter.',
            ],
          ] as [keyof TConnectionValues, string, string, string?][])
        : ([
            [
              'role_name',
              'IAM role name',
              'nuon-cloud-connection',
              'Use a new or existing IAM role in this AWS account.',
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
      <SectionHeader
        title={imageRepository ? 'New pull identity' : 'Connection details'}
        description={
          imageRepository
            ? 'Create an identity for pulling images. Nuon manages it as a cloud connection; no long-lived credentials are stored.'
            : isImage
              ? 'Nuon saves this identity and registry target. Repository permissions are configured separately.'
              : 'Nuon saves this identity. Review and apply its cloud setup after creation.'
        }
      />
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
            labelProps={{
              labelText: imageRepository ? 'Identity name' : 'Connection name',
            }}
            placeholder="acme-images"
            className="!font-mono"
            disabled={isPending}
          />
        )}
      </form.Field>
      {!imageRepository && !isGCP && !isAzure && (
        <form.Field name="preset">
          {(field) => (
            <FormSelect
              field={field}
              id="connection-access"
              labelProps={{ labelText: 'Access type' }}
              disabled={isPending}
              options={[
                { value: 'ecr', label: 'ECR image imports' },
                { value: 'stacks', label: 'Manage install stacks' },
                { value: 'custom', label: 'Custom policy' },
              ]}
            />
          )}
        </form.Field>
      )}
      <section
        className="flex flex-col gap-4 border-t pt-6"
        aria-label="Cloud identity"
      >
        <SectionHeader
          title="Cloud identity"
          description={`The identity Nuon will use in ${isAzure ? 'Azure' : isGCP ? 'Google Cloud' : 'AWS'}.`}
        />
        <form.Field
          name="principal_choice"
          listeners={{
            onChange: ({ value }) => {
              const selected = principals.find((c) => c.id === value)
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
                  : defaultValues.role_name
              )
              form.setFieldValue(
                'service_account_name',
                selected
                  ? selected.principal.split('@')[0]
                  : defaultValues.service_account_name
              )
              form.setFieldValue(
                'gcp_project_number',
                selected?.gcp_project_number || defaultValues.gcp_project_number
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
                form.setFieldValue('azure_resource_group', registry?.[1] || '')
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
              id="connection-principal-choice"
              labelProps={{ labelText: 'Principal' }}
              placeholder="Choose a principal or add a new one"
              helperText="Choose an existing cloud principal or enter details for another one. Its cloud permissions apply to each connection using it."
              disabled={isPending}
              searchable
              options={[
                ...principals.map((connection) => ({
                  value: connection.id,
                  label: connection.principal,
                  content: (
                    <span className="flex min-w-0 flex-col">
                      <Text family="mono" className="break-all">
                        {connection.principal}
                      </Text>
                      {platform === 'azure' && (
                        <Text variant="subtext" theme="neutral">
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
                })),
                { value: 'new', label: 'Add new principal' },
              ]}
            />
          )}
        </form.Field>
        {selectedPrincipal && !isAzure && (
          <LabeledValue label={isGCP ? 'Project ID' : 'AWS account ID'}>
            <Text family="mono">{selectedPrincipal.target_id}</Text>
          </LabeledValue>
        )}
        {identityFields
          .filter(
            ([name]) =>
              principalChoice === 'new' ||
              (selectedPrincipal && isAzure && name === 'target_id')
          )
          .map(([name, label, placeholder, helperText]) => (
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
        <PrincipalSharing connections={shared} orgId={orgId} />
      </section>
      {isImage && (
        <section
          className="flex flex-col gap-4 border-t pt-6"
          aria-label="Registry target"
        >
          <SectionHeader
            title="Registry target"
            description="The registry coordinates this identity can be assigned to."
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
          Creating a connection does not change cloud permissions.
        </Text>
        <Button
          type="submit"
          variant="primary"
          disabled={!canSubmit || isPending}
          tooltipProps={
            !canSubmit && !isPending
              ? {
                  tipContent:
                    'Cannot create connection — complete the required details',
                }
              : undefined
          }
        >
          {isPending ? (
            <>
              <Icon variant="Loading" /> Creating connection
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
