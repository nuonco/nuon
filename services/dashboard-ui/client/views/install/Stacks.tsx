import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { InstallStacksTable, InstallStacksTableSkeleton } from '@/components/stacks/InstallStacksTable'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { getAppConfig, getAppConfigs } from '@/lib'
import { hasNewerAppConfig, hasStackConfigChanged } from '@/utils/app-utils'

export const Stacks = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: configResult, isLoading: isLoadingConfig } = useQuery({
    queryKey: [
      'app-config',
      org?.id,
      install?.app_id,
      install?.app_config_id,
      'recurse',
    ],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_config_id,
  })

  const config = configResult

  const { data: latestConfigs } = useQuery({
    queryKey: ['app-configs', org?.id, install?.app_id, 'latest'],
    queryFn: () =>
      getAppConfigs({ orgId: org.id, appId: install.app_id, limit: 1, offset: 0 }),
    enabled: !!org?.id && !!install?.app_id,
  })
  const latestConfigSummary = latestConfigs?.[0]
  const newerAppConfig = hasNewerAppConfig(latestConfigSummary, install)

  const { data: latestFullConfig } = useQuery({
    queryKey: ['app-config', org?.id, install?.app_id, latestConfigSummary?.id, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: latestConfigSummary!.id!,
        recurse: true,
      }),
    enabled: newerAppConfig && !!latestConfigSummary?.id,
  })

  const stackChanged = hasStackConfigChanged(config, latestFullConfig)

  return (
    <PageSection>
      <PageTitle title={`Stacks | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/stacks`,
            text: 'Stacks',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install stacks
        </Text>
        <Text variant="subtext" theme="neutral">
          View your install stack config and versions below.
        </Text>
      </HeadingGroup>

      {stackChanged && (
        <Banner theme="info">
          <div className="flex items-center gap-8">
            <div className="flex flex-col">
              <Text weight="strong">New stack config available</Text>
              <Text variant="subtext" theme="neutral">
                A newer stack config (v{latestFullConfig?.version}) is available. This install
                is using v{config?.version}.
              </Text>
            </div>
            <Button
              className="ml-auto"
              href={`/${org.id}/apps/${install.app_id}`}
              variant="secondary"
            >
              View latest config
            </Button>
          </div>
        </Banner>
      )}

      {(() => {
        const installConfig = install?.install_config

        const isRunnerOverridden = Boolean(installConfig?.runner_nested_template_url)
        const effectiveRunnerURL = installConfig?.runner_nested_template_url || config?.stack?.runner_nested_template_url

        const isVpcOverridden = Boolean(installConfig?.vpc_nested_template_url)
        const effectiveVpcURL = installConfig?.vpc_nested_template_url || config?.stack?.vpc_nested_template_url

        const appStacks = config?.stack?.custom_nested_stacks || []
        const installStacks = installConfig?.custom_nested_stacks || []
        const overrideMap = new Map(installStacks.map((s) => [s.name, s]))
        const seen = new Set<string>()
        const effectiveStacks = [
          ...appStacks.map((s) => {
            seen.add(s.name!)
            if (overrideMap.has(s.name!)) {
              return { ...overrideMap.get(s.name!)!, _status: 'overridden' as const }
            }
            return { ...s, _status: 'default' as const }
          }),
          ...installStacks
            .filter((s) => !seen.has(s.name!))
            .map((s) => ({ ...s, _status: 'new' as const })),
        ]

        if (isLoadingConfig) {
          return (
            <Card>
              <Skeleton width="135px" height="24px" />
              <div className="grid grid-cols-6 gap-3">
                <LabeledValue label={<Skeleton height="17px" width="100px" />}>
                  <Skeleton height="17px" width="20px" />
                </LabeledValue>
                <LabeledValue label={<Skeleton height="17px" width="60px" />}>
                  <Skeleton height="17px" width="110px" />
                </LabeledValue>
                <LabeledValue label={<Skeleton height="17px" width="65px" />}>
                  <Skeleton height="17px" width="180px" />
                </LabeledValue>
              </div>
            </Card>
          )
        }

        if (!config) {
          return (
            <Card>
              <EmptyState
                emptyTitle="No stack config"
                emptyMessage="Unable to load this install stack config"
                variant="table"
              />
            </Card>
          )
        }

        return (
          <Card>
            <Text weight="strong">Current stack config</Text>
            <div className="grid grid-cols-6 gap-3">
              <LabeledValue label="App config version">
                {config?.version?.toString()}
              </LabeledValue>
              <LabeledValue label="Stack type">
                {config?.stack?.type}
              </LabeledValue>
              <LabeledValue label="Stack name">
                {config?.stack?.name}
              </LabeledValue>
              {effectiveRunnerURL ? (
                <LabeledValue
                  className="col-span-6"
                  label={
                    <span className="flex items-center gap-2">
                      <Text variant="subtext" theme="neutral">Runner nested template URL</Text>
                      {isRunnerOverridden && <Badge size="sm" theme="info">install override</Badge>}
                    </span>
                  }
                >
                  <Text variant="subtext">
                    <Link href={effectiveRunnerURL} isExternal>
                      {effectiveRunnerURL}
                    </Link>
                  </Text>
                </LabeledValue>
              ) : null}
              {effectiveVpcURL ? (
                <LabeledValue
                  className="col-span-6"
                  label={
                    <span className="flex items-center gap-2">
                      <Text variant="subtext" theme="neutral">VPC nested template URL</Text>
                      {isVpcOverridden && <Badge size="sm" theme="info">install override</Badge>}
                    </span>
                  }
                >
                  <Text variant="subtext">
                    <Link href={effectiveVpcURL} isExternal>
                      {effectiveVpcURL}
                    </Link>
                  </Text>
                </LabeledValue>
              ) : null}
            </div>
            {effectiveStacks.length > 0 && (
              <div className="flex flex-col gap-2 mt-2">
                <Text variant="subtext" weight="strong">Custom nested stacks</Text>
                <div className="flex flex-col gap-1">
                  {effectiveStacks.map((s) => (
                    <div key={s.name} className="flex items-center gap-3 text-sm">
                      <Text variant="subtext" family="mono">{s.name}</Text>
                      <Text variant="subtext">
                        <Link href={s.template_url!} isExternal>{s.template_url}</Link>
                      </Text>
                      {s._status === 'overridden' && <Badge size="sm" theme="info">install override</Badge>}
                      {s._status === 'new' && <Badge size="sm" theme="success">install-only</Badge>}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </Card>
        )
      })()}

      <div className="flex flex-col gap-4">
        <Text weight="strong">Install stack versions</Text>
        <InstallStacksTable shouldPoll />
      </div>
    </PageSection>
  )
}
