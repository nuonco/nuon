import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { SandboxRunsTimeline } from '@/components/sandbox/SandboxRunsTimeline'
import { ManagementDropdown } from '@/components/sandbox/management/ManagementDropdown'
import { SandboxRunConfigCard } from '@/components/sandbox/SandboxRunConfigCard'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getInstallSandboxRuns } from '@/lib'

const LIMIT = 10

export const Sandbox = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: configResult } = useQuery({
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

  const { data: runsResult } = useQuery({
    queryKey: ['install-sandbox-runs', org?.id, install?.id, offset],
    queryFn: () =>
      getInstallSandboxRuns({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT + 1,
        offset,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const allRuns = runsResult ?? []
  const hasNext = allRuns.length > LIMIT
  const runs = allRuns.slice(0, LIMIT)
  const pagination = { hasNext, offset, limit: LIMIT }

  const sandboxConfig = configResult?.sandbox

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/sandbox`,
            text: 'Sandbox',
          },
        ]}
      />

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-y md:divide-x">
        <div className="md:col-span-8 divide-y flex-auto flex flex-col">
          {sandboxConfig ? (
            <div className="p-6">
              <SandboxRunConfigCard config={sandboxConfig} />
            </div>
          ) : null}
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <div className="p-6 flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <HeadingGroup>
                <Text variant="base" weight="strong">
                  Sandbox controls
                </Text>
              </HeadingGroup>
              <ManagementDropdown />
            </div>
          </div>

          <div className="p-6 flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Sandbox history
            </Text>
            <SandboxRunsTimeline
              initRuns={runs}
              pagination={pagination}
              shouldPoll
            />
          </div>
        </div>
      </div>
    </PageSection>
  )
}
