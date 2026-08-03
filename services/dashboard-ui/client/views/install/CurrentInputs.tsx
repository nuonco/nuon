import { useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { LabeledValue } from '@/components/common/LabeledValue'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { EditInputsButton } from '@/components/installs/management/EditInputs'
import { InputValue } from '@/components/installs/management/InputValue'
import {
  InputsFilterBar,
  InputsNoResults,
} from '@/components/installs/InputsFilter'
import { ComponentOverridesList } from '@/components/install-overrides/ComponentOverridesList'
import {
  useInputsFilter,
  type TInputsFilterGroup,
} from '@/hooks/use-inputs-filter'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getInstallCurrentInputs } from '@/lib'
import { normalizeAppInputGroups } from '@/utils/app-utils'
import {
  COMPONENT_OVERRIDE_INPUT_GROUP,
  getInputDisplayName,
} from '@/utils/install-utils'

// Mirrors app.CloudPlatformMetadata in ctl-api — the generated type is an
// opaque object because the field is serialized as a plain JSON object
type TCloudPlatformMetadata = {
  target_account_id?: string
  observed_account_id?: string
  target_project_id?: string
  observed_project_id?: string
  target_subscription_id?: string
  observed_subscription_id?: string
  target_source?: string
}

const CLOUD_METADATA_LABELS: Array<{
  key: keyof TCloudPlatformMetadata
  label: string
}> = [
  { key: 'target_account_id', label: 'Target AWS account' },
  { key: 'observed_account_id', label: 'Observed AWS account' },
  { key: 'target_project_id', label: 'Target GCP project' },
  { key: 'observed_project_id', label: 'Observed GCP project' },
  { key: 'target_subscription_id', label: 'Target Azure subscription' },
  { key: 'observed_subscription_id', label: 'Observed Azure subscription' },
  { key: 'target_source', label: 'Target source' },
]

export const CurrentInputs = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const phoneHomeAuthEnabled = !!org?.features?.['phone-home-auth']
  const cloudMetadata = (install?.cloud_platform_metadata ??
    {}) as TCloudPlatformMetadata
  const cloudMetadataRows = CLOUD_METADATA_LABELS.filter(
    ({ key }) => cloudMetadata[key]
  )
  const hasCloudIdentifier = cloudMetadataRows.some(
    ({ key }) => key !== 'target_source'
  )
  const phoneHomeAuth = install?.phone_home_auth

  const { data: inputs, isLoading: inputsLoading } = useQuery({
    queryKey: ['install-inputs', org?.id, install?.id],
    queryFn: () =>
      getInstallCurrentInputs({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const { appConfig: config, isLoading: configLoading } = useInstallAppConfig()

  const isLoading = inputsLoading || configLoading
  const redacted = inputs?.redacted_values ?? {}
  const hasInputs = Object.keys(redacted).length > 0
  const inputGroups = config
    ? normalizeAppInputGroups(
        config.input?.input_groups ?? [],
        config.input?.inputs ?? []
      )
    : []
  const hasConfig = inputGroups.length > 0

  const {
    search,
    setSearch,
    attributeFilters,
    sourceFilters,
    setAttributeFilters,
    setSourceFilters,
    toggleAttribute,
    toggleSource,
    clearAllFilters,
    clearAll,
    filterCount,
    hasActiveSearch,
    hasActiveFilters,
    filteredGroups,
    filteredFlatInputs,
  } = useInputsFilter({
    inputGroups: inputGroups as TInputsFilterGroup[],
    redacted,
  })

  const noResultsEmpty = (
    <InputsNoResults
      search={search}
      hasActiveSearch={hasActiveSearch}
      hasActiveFilters={hasActiveFilters}
      onClearSearch={() => setSearch('')}
      onClearFilters={clearAllFilters}
      onClearAll={clearAll}
    />
  )

  return (
    <PageSection>
      <PageTitle title={`Current inputs | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/inputs`,
            text: 'Current inputs',
          },
        ]}
      />
      {phoneHomeAuthEnabled && !hasCloudIdentifier && (
        <Banner theme="warn">
          <div className="flex flex-col">
            <Text weight="strong">No cloud platform metadata</Text>
            <Text variant="subtext" theme="neutral">
              Phone-home auth is enabled for this org, but this install has no
              target or observed cloud account recorded.
            </Text>
          </div>
        </Banner>
      )}

      <div className="flex items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Current inputs
          </Text>
          <Text variant="subtext" theme="neutral">
            The current input values for this install.
          </Text>
        </HeadingGroup>
        <div className="shrink-0">
          <EditInputsButton variant="secondary" />
        </div>
      </div>

      {phoneHomeAuthEnabled && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {hasCloudIdentifier && (
            <Card className="flex flex-col gap-4 h-full">
              <Text weight="strong">Cloud Platform Metadata</Text>
              <div className="flex flex-wrap gap-6 items-start">
                {cloudMetadataRows.map(({ key, label }) => (
                  <LabeledValue key={key} label={label}>
                    <Text family="mono" variant="subtext">
                      {cloudMetadata[key]}
                    </Text>
                  </LabeledValue>
                ))}
              </div>
            </Card>
          )}
          <Card className="flex flex-col gap-4 h-full">
            <Text weight="strong">Phone Home Auth</Text>
            {phoneHomeAuth ? (
              <div className="flex flex-wrap gap-6 items-start">
                <LabeledValue label="Provisioned">
                  <Time
                    format="relative"
                    variant="subtext"
                    time={phoneHomeAuth.provisioned_at}
                  />
                </LabeledValue>
                <LabeledValue label="Last verified">
                  {phoneHomeAuth.last_verified_at ? (
                    <Time
                      format="relative"
                      variant="subtext"
                      time={phoneHomeAuth.last_verified_at}
                    />
                  ) : (
                    <Text variant="subtext" theme="neutral">
                      —
                    </Text>
                  )}
                </LabeledValue>
                <LabeledValue label="Last rejected">
                  {phoneHomeAuth.last_rejected_at ? (
                    <Time
                      format="relative"
                      variant="subtext"
                      time={phoneHomeAuth.last_rejected_at}
                    />
                  ) : (
                    <Text variant="subtext" theme="neutral">
                      —
                    </Text>
                  )}
                </LabeledValue>
              </div>
            ) : (
              <Text variant="subtext" theme="neutral">
                Phone-home credentials have not been provisioned for this
                install yet.
              </Text>
            )}
          </Card>
        </div>
      )}

      {!isLoading && (hasConfig || hasInputs) ? (
        <div className="flex justify-end">
          <InputsFilterBar
            search={search}
            onSearchChange={setSearch}
            showFilters={hasConfig}
            attributeFilters={attributeFilters}
            sourceFilters={sourceFilters}
            setAttributeFilters={setAttributeFilters}
            setSourceFilters={setSourceFilters}
            toggleAttribute={toggleAttribute}
            toggleSource={toggleSource}
            clearAllFilters={clearAllFilters}
            filterCount={filterCount}
            hasActiveFilters={hasActiveFilters}
          />
        </div>
      ) : null}

      {isLoading ? (
        <Skeleton height="200px" width="100%" />
      ) : hasConfig ? (
        filteredGroups.length === 0 ? (
          noResultsEmpty
        ) : (
          <div className="flex flex-col gap-4">
            {filteredGroups.map((group) => {
              const groupInputs = group.app_inputs ?? []
              if (groupInputs.length === 0) return null

              return (
                <Expand
                  isOpen
                  id={group.id}
                  key={group.id}
                  heading={
                    <div className="flex flex-col items-start">
                      <Text weight="strong">{group.display_name}</Text>
                      {group.description && (
                        <Text variant="subtext" theme="neutral">
                          {group.description}
                        </Text>
                      )}
                    </div>
                  }
                  className="border rounded-md"
                  headerClassName="!px-4"
                >
                  {group.name === COMPONENT_OVERRIDE_INPUT_GROUP ? (
                    <div className="p-4 border-t bg-black/[0.0075] dark:bg-white/[0.0075]">
                      <ComponentOverridesList
                        inputs={groupInputs}
                        values={redacted}
                      />
                    </div>
                  ) : (
                    <div className="p-4 border-t bg-black/[0.0075] dark:bg-white/[0.0075]">
                      <PropertyGrid
                        align="start"
                        columns={[
                          { key: 'name', header: 'Name' },
                          { key: 'value', header: 'Current value' },
                          { key: 'default', header: 'Default' },
                        ]}
                        gridTemplate="minmax(150px, 1fr) minmax(150px, 2fr) minmax(120px, 1fr)"
                        values={groupInputs.map((input) => ({
                          name: (
                            <span className="flex flex-col">
                              <Text variant="subtext" weight="strong">
                                {input.display_name}
                              </Text>
                              <Text
                                variant="label"
                                family="mono"
                                theme="neutral"
                              >
                                {input.name
                                  ? getInputDisplayName(input.name)
                                  : null}
                              </Text>
                            </span>
                          ),
                          value: (
                            <InputValue
                              name={input.name}
                              value={
                                input.name ? redacted[input.name] : undefined
                              }
                            />
                          ),
                          default: (
                            <Text variant="label" family="mono" theme="neutral">
                              {input?.default}
                            </Text>
                          ),
                        }))}
                      />
                    </div>
                  )}
                </Expand>
              )
            })}
          </div>
        )
      ) : hasInputs ? (
        filteredFlatInputs.length === 0 ? (
          noResultsEmpty
        ) : (
          <PropertyGrid
            values={filteredFlatInputs.map(([key, value]) => ({
              key: getInputDisplayName(key),
              value: <InputValue name={key} value={String(value)} />,
            }))}
          />
        )
      ) : (
        <EmptyState
          emptyTitle="No inputs configured"
          emptyMessage="This install doesn't have any inputs yet. Use the manage menu to edit inputs."
          variant="diagram"
          size="sm"
        />
      )}
    </PageSection>
  )
}
