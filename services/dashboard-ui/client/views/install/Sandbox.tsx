import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { DriftedBanner } from '@/components/install-components/DriftedBanner'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { SandboxRunsTimeline } from '@/components/sandbox/SandboxRunsTimeline'
import { ManagementDropdown } from '@/components/sandbox/management/ManagementDropdown'
import {
  SandboxConfigCard,
  SandboxConfigCardSkeleton,
} from '@/components/sandbox/SandboxConfigCard'
import { TerraformWorkspaceCard } from '@/components/terraform-workspace/TerraformWorkspaceCard'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Panel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfig } from '@/lib'

export const Sandbox = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addPanel } = useSurfaces()

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

  const latestSandboxRunId = install?.install_sandbox_runs?.at(0)?.id
  const driftedObject = install?.drifted_objects?.find(
    (d) =>
      d?.target_type === 'install_sandbox_run' &&
      d?.target_id === latestSandboxRunId
  )

  return (
    <PageSection>
      <PageTitle title={`Sandbox | ${install?.name}`} />
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

      <div className="@container flex flex-col flex-auto gap-6">
        <div className="flex items-start justify-between">
          <HeadingGroup>
            <Text variant="base" weight="strong">
              Sandbox details
            </Text>
            <ID>{install?.sandbox?.id}</ID>
          </HeadingGroup>
          <div className="flex items-center gap-2">
            <div className="@5xl:hidden">
              <Button
                variant="secondary"
                onClick={() =>
                  addPanel(
                    <Panel heading="Sandbox history">
                      <SandboxRunsTimeline shouldPoll />
                    </Panel>
                  )
                }
              >
                <Icon variant="ClockCounterClockwiseIcon" size={16} />
                Sandbox history
              </Button>
            </div>
            <ManagementDropdown />
          </div>
        </div>

        <div className="grid grid-cols-1 @5xl:grid-cols-12 gap-6">
          <div className="@5xl:col-span-8 flex flex-col gap-6 min-w-0">
            {driftedObject ? <DriftedBanner drifted={driftedObject} /> : null}

            {sandboxConfig ? (
              <SandboxConfigCard config={sandboxConfig} />
            ) : (
              <SandboxConfigCardSkeleton />
            )}

            <TerraformWorkspaceCard />
          </div>

          <div className="hidden @5xl:flex flex-col @5xl:col-span-4 gap-4 min-w-0">
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
