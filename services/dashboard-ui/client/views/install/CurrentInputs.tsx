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
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getInstallCurrentInputs } from '@/lib'
import { normalizeAppInputGroups } from '@/utils/app-utils'

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

  const isLoading = inputsLoading || configLoading
  const redacted = inputs?.redacted_values ?? {}
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
      <div className="flex items-start justify-between">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Current inputs
          </Text>
          <Text variant="subtext" theme="neutral">
            The current input values for this install.
          </Text>
        </HeadingGroup>
        <div className="flex items-center gap-2">
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
                              {input.name}
                            </Text>
                          </span>
                        ),
                        value:
                          input.name && redacted[input.name] != null ? (
                            String(redacted[input.name]) === '' ? (
                              <Text
                                variant="subtext"
                                family="mono"
                                theme="neutral"
                              >
                                &quot;&quot;
                              </Text>
                            ) : (
                              <Text
                                variant="subtext"
                                family="mono"
                                weight="strong"
                              >
                                {String(redacted[input.name])}
                              </Text>
                            )
                          ) : (
                            <Text
                              variant="subtext"
                              family="mono"
                              theme="neutral"
                            >
                              —
                            </Text>
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
            key,
            value,
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
