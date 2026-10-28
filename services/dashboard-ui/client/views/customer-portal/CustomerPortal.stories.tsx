export default {
  title: 'Playground / Customer portal / Install',
  meta: { fullBleed: true },
}

import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { EmptyState } from '@/components/common/EmptyState'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { CustomerInstallShell } from '@/components/shared-installs/CustomerInstallShell/CustomerInstallShell'
import { SharedInstallsStoryChrome } from '@/components/shared-installs/story-chrome'
import type { TInstallShareScope } from '@/components/shared-installs/types'

const BASE_PATH = '/org-1/shared-installs/inst-1'
const ALL_SCOPES: TInstallShareScope[] = [
  'approvals',
  'readme',
  'resources',
  'operations',
]

const readmeMarkdown = `# Acme production

Welcome to your install.

## Getting started

1. Review pending approvals
2. Check deployment history
3. Inspect resources when something looks off
`

const deploymentEvents = [
  {
    created_at: '2026-10-27T09:30:00Z',
    title: 'Deploy v1.4.3',
    status: 'in-progress',
    caption: 'Applying component updates',
  },
  {
    created_at: '2026-10-26T12:00:00Z',
    title: 'Deploy v1.4.2',
    status: 'success',
    caption: 'Completed successfully',
  },
]

const operationEvents = [
  {
    created_at: '2026-10-25T16:00:00Z',
    title: 'Sync secrets',
    status: 'success',
    caption: 'Secrets synced to the cluster',
  },
  {
    created_at: '2026-10-24T11:20:00Z',
    title: 'Ad hoc action',
    status: 'failed',
    caption: 'Action exited with an error',
  },
]

const PortalPage = ({
  scopes,
  title,
  description,
  children,
}: {
  scopes: TInstallShareScope[]
  title: string
  description?: string
  children: ReactNode
}) => (
  <SharedInstallsStoryChrome
    breadcrumbLinks={[
      { path: '/org-1', text: 'Acme' },
      { path: '/org-1/shared-installs', text: 'Shared installs' },
      { path: BASE_PATH, text: 'Acme production' },
    ]}
  >
    <CustomerInstallShell
      installName="Acme production"
      installId="instxxxxxxxxxxxxxxxxxxxxxxxx"
      basePath={BASE_PATH}
      scopes={scopes}
    >
      <PageSection>
        <SectionHeader title={title} description={description} />
        {children}
      </PageSection>
    </CustomerInstallShell>
  </SharedInstallsStoryChrome>
)

export const Overview = () => (
  <PortalPage
    scopes={ALL_SCOPES}
    title="Overview"
    description="Status and recent activity for this install."
  >
    <div className="flex flex-col gap-4">
      <Text variant="body">
        This install is healthy. Latest deploy succeeded 2 days ago.
      </Text>
      <EmptyState
        emptyTitle="No open issues"
        emptyMessage="Resources and operations look healthy."
      />
    </div>
  </PortalPage>
)

export const Deployments = () => (
  <PortalPage
    scopes={ALL_SCOPES}
    title="Deployments"
    description="Deployment history for this install."
  >
    <Timeline
      events={deploymentEvents}
      pagination={{ limit: 10, offset: 0, hasNext: false }}
      renderEvent={(event) => (
        <TimelineEvent
          createdAt={event.created_at}
          caption={event.caption}
          status={event.status}
          title={event.title}
        />
      )}
    />
  </PortalPage>
)

export const Resources = () => (
  <PortalPage
    scopes={ALL_SCOPES}
    title="Resources"
    description="Resources provisioned for this install."
  >
    <EmptyState
      emptyTitle="Resources in sync"
      emptyMessage="No drifted resources for this install."
    />
  </PortalPage>
)

export const Operations = () => (
  <PortalPage
    scopes={ALL_SCOPES}
    title="Operations"
    description="Recent operations for this install."
  >
    <Timeline
      events={operationEvents}
      pagination={{ limit: 10, offset: 0, hasNext: false }}
      renderEvent={(event) => (
        <TimelineEvent
          createdAt={event.created_at}
          caption={event.caption}
          status={event.status}
          title={event.title}
        />
      )}
    />
  </PortalPage>
)

export const Readme = () => (
  <PortalPage scopes={ALL_SCOPES} title="README">
    <Markdown content={readmeMarkdown} />
  </PortalPage>
)

export const Approvals = () => (
  <PortalPage
    scopes={ALL_SCOPES}
    title="Approvals"
    description="Pending plans that need a response."
  >
    <Banner theme="warn">
      <span className="flex flex-col gap-1">
        <Text weight="strong">Terraform plan requires review</Text>
        <Text variant="subtext">
          This Terraform plan is ready for review. Inspect the proposed
          infrastructure changes before applying them.
        </Text>
      </span>
    </Banner>
  </PortalPage>
)

export const LimitedScopes = () => (
  <PortalPage
    scopes={['readme', 'resources']}
    title="README"
    description="Only README and resources were shared with this customer."
  >
    <Markdown content={readmeMarkdown} />
  </PortalPage>
)
