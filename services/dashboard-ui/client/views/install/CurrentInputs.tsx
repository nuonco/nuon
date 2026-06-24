import { useQuery } from '@tanstack/react-query'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { EditInputsButton } from '@/components/installs/management/EditInputs'
import { InputValue } from '@/components/installs/management/InputValue'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import {
  getAppConfig,
  getInstallComponents,
  getInstallCurrentInputs,
} from '@/lib'
import { normalizeAppInputGroups } from '@/utils/app-utils'
import {
  disabledToggleableDeps,
  getEnabledOverrideComponent,
  getInputDisplayName,
} from '@/utils/install-utils'

export const CurrentInputs = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: inputs, isLoading: inputsLoading } = useQuery({
    queryKey: ['install-inputs', org?.id, install?.id],
    queryFn: () =>
      getInstallCurrentInputs({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const { data: config, isLoading: configLoading } = useQuery({
    queryKey: ['app-config', org?.id, install?.app_id, install?.app_config_id],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!install?.app_config_id,
  })

  const { data: componentsResult } = useQuery({
    queryKey: ['install-components', org?.id, install?.id],
    queryFn: () =>
      getInstallComponents({ orgId: org.id, installId: install.id, limit: 100 }),
    enabled: !!org?.id && !!install?.id,
  })

  const isLoading = inputsLoading || configLoading
  const redacted = inputs?.redacted_values ?? {}
  const configComponents = config?.component_config_connections ?? []
  const effectiveEnabledByName: Record<string, boolean | undefined> = {}
  for (const c of componentsResult?.data ?? []) {
    if (c.component?.name) effectiveEnabledByName[c.component.name] = c.enabled
  }
  const hasInputs = Object.keys(redacted).length > 0
  const inputGroups = config
    ? normalizeAppInputGroups(
        config.input?.input_groups ?? [],
        config.input?.inputs ?? []
      )
    : []

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

      {isLoading ? (
        <Skeleton height="200px" width="100%" />
      ) : inputGroups.length > 0 ? (
        <div className="flex flex-col gap-4">
          {(inputGroups as any[]).map((group) => {
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
                <div className="p-4 border-t bg-code">
                  <PropertyGrid
                    columns={[
                      { key: 'name', header: 'Name' },
                      { key: 'value', header: 'Current value' },
                      { key: 'default', header: 'Default' },
                    ]}
                    gridTemplate="minmax(150px, 1fr) minmax(150px, 2fr) minmax(120px, 1fr)"
                    values={groupInputs.map(
                      (input: {
                        name?: string
                        display_name?: string
                        default?: string
                      }) => ({
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
                            {(() => {
                              const comp = input.name
                                ? getEnabledOverrideComponent(input.name)
                                : null
                              if (!comp) return null
                              const own =
                                input.name &&
                                String(redacted[input.name]) === 'true'
                              if (!own) return null
                              if (effectiveEnabledByName[comp] !== false)
                                return null
                              const blockers = disabledToggleableDeps(
                                comp,
                                configComponents,
                                effectiveEnabledByName
                              )
                              return (
                                <Text
                                  variant="label"
                                  theme="warn"
                                  className="mt-0.5"
                                >
                                  {blockers.length > 0
                                    ? `Effectively disabled — requires ${blockers.join(', ')}`
                                    : 'Effectively disabled — a required component is turned off'}
                                </Text>
                              )
                            })()}
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
                      })
                    )}
                  />
                </div>
              </Expand>
            )
          })}
        </div>
      ) : hasInputs ? (
        <PropertyGrid
          values={Object.entries(redacted).map(([key, value]) => ({
            key: getInputDisplayName(key),
            value: <InputValue name={key} value={String(value)} />,
          }))}
        />
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
