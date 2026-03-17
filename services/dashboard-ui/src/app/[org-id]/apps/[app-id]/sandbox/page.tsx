import type { Metadata } from 'next'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { AppSandbox as SandboxConfig } from '@/components/apps/config/AppSandbox'
import { BuildSandboxButton } from '@/components/sandbox/management/BuildSandbox'
import { getApp, getAppConfig, getAppConfigs, getOrg } from '@/lib'
import type { TPageProps } from '@/types'

type TAppPageProps = TPageProps<'org-id' | 'app-id'>

export async function generateMetadata({
  params,
}: TAppPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getApp({ appId, orgId })

  return {
    title: `Sandbox | ${app.name} | Nuon`,
  }
}

async function SandboxContent({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) {
  const { data: configs } = await getAppConfigs({ appId, orgId, limit: 1 })
  const appConfigId = configs?.at(0)?.id

  if (!appConfigId) {
    return (
      <Card className="h-full">
        <EmptyState
          variant="diagram"
          emptyTitle="No app configuration"
          emptyMessage="Configure your application to see sandbox details here."
        />
      </Card>
    )
  }

  const { data: config, error } = await getAppConfig({
    appConfigId,
    appId,
    orgId,
    recurse: true,
  })

  if (error || !config?.sandbox) {
    return (
      <Card className="h-full">
        <EmptyState
          variant="diagram"
          emptyTitle="No sandbox configuration"
          emptyMessage="Configure a sandbox in your application configuration to see it here."
        />
      </Card>
    )
  }

  return (
    <Card className="h-fit flex flex-col gap-4">
      <Text weight="strong">Sandbox config</Text>
      <SandboxConfig appConfig={config} />
    </Card>
  )
}

export default async function AppSandboxPage({ params }: TAppPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const [{ data: app }, { data: org }] = await Promise.all([
    getApp({ appId, orgId }),
    getOrg({ orgId }),
  ])

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name,
          },
          {
            path: `/${orgId}/apps/${appId}/sandbox`,
            text: 'Sandbox',
          },
        ]}
      />
      <div className="flex justify-between items-center">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Sandbox
          </Text>
        </HeadingGroup>
        <BuildSandboxButton variant="primary" />
      </div>

      <AsyncBoundary
        errorFallback={
          <Card className="h-full">
            <EmptyState
              variant="diagram"
              emptyTitle="Unable to load sandbox"
              emptyMessage="An error occurred loading the sandbox configuration."
            />
          </Card>
        }
        loadingFallback={
          <Card className="h-full flex flex-col gap-4">
            <Text weight="strong">Loading sandbox config...</Text>
          </Card>
        }
      >
        <SandboxContent appId={appId} orgId={orgId} />
      </AsyncBoundary>
    </PageSection>
  )
}