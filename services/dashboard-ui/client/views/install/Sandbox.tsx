import { useQuery } from '@tanstack/react-query'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { SandboxRunsTimeline } from '@/components/sandbox/SandboxRunsTimeline'
import { ManagementDropdown } from '@/components/sandbox/management/ManagementDropdown'
import { SandboxRunConfigCard } from '@/components/sandbox/SandboxRunConfigCard'
import { TerraformWorkspaceCard } from '@/components/sandbox/TerraformWorkspaceCard'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig } from '@/lib'

export const Sandbox = () => {
  const { org } = useOrg()
  const { install } = useInstall()

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
            <div className="p-6 flex flex-col gap-4">
              <div className="flex items-center justify-between">
                <Text variant="base" weight="strong">
                  Config
                </Text>
                <Text variant="subtext">
                  <Link href={`/${org?.id}/apps/${install?.app_id}`}>
                    Details
                    <Icon variant="CaretRightIcon" />
                  </Link>
                </Text>
              </div>
              <SandboxRunConfigCard config={sandboxConfig} />
            </div>
          ) : null}

          <div className="p-6">
            <TerraformWorkspaceCard />
          </div>
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <div className="p-6 flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <Text variant="base" weight="strong">
                Sandbox controls
              </Text>
              <ManagementDropdown />
            </div>
          </div>

          <div className="p-6 flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Sandbox history
            </Text>
            <SandboxRunsTimeline shouldPoll />
          </div>
        </div>
      </div>
    </PageSection>
  )
}
