import {
  BranchActivityFeed,
  type TBranchActivityItem,
} from '@/components/orgs/BranchActivityFeed'
import { AnnouncementsList } from '@/components/orgs/AnnouncementsList'
import type { IAnnouncement } from '@/components/orgs/AnnouncementCard'
import { MainLayout } from '@/components/layout/MainLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageGrid } from '@/components/layout/PageGrid'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TNuonVersion } from '@/types'

export interface IDashboardLandingPage {
  orgName: string
  runs: TBranchActivityItem[]
  announcements: IAnnouncement[]
}

const playgroundVersions: TNuonVersion = {
  api: { git_ref: 'abc1234', version: '1.2.3' },
  ui: { version: '4.5.6' },
}

export const DashboardLandingPage = ({
  orgName,
  runs,
  announcements,
}: IDashboardLandingPage) => (
  <MainLayout versions={playgroundVersions}>
    <PageLayout>
      <SectionHeader
        variant="page"
        title={`Welcome to ${orgName}`}
        description="Manage your applications and deployed installs."
      />
      <PageContent className="@container">
        <PageGrid className="@4xl:divide-x flex-auto !grid-cols-1 @4xl:!grid-cols-[1fr_400px]">
          <PageSection className="flex-1 @4xl:border-r">
            <SectionHeader
              title="Recent updates"
              description="Applied and in-flight updates across your app branches, and the installs each one updated."
            />
            <BranchActivityFeed items={runs} />
          </PageSection>
          <PageSection className="w-full hidden @4xl:flex">
            <AnnouncementsList
              announcements={announcements}
              disableDismissMemory
            />
          </PageSection>
        </PageGrid>
      </PageContent>
    </PageLayout>
  </MainLayout>
)
