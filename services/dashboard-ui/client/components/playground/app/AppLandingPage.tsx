import {
  BranchCardsComponent,
  type TBranchCardData,
} from '@/components/branches/BranchCards'
import { Button } from '@/components/common/Button'
import {
  AppInstallsList,
  type TAppInstallListItem,
} from '@/components/installs/AppInstallsList'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { PageContent } from '@/components/layout/PageContent'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { AppSourceChipComponent } from '@/components/apps/AppSourceChip'

export interface IAppLandingPage {
  app: {
    id: string
    name: string
    repo: string
    repoHref?: string
  }
  branches: TBranchCardData[]
  installs: TAppInstallListItem[]
}

export const AppLandingPage = ({
  app,
  branches,
  installs,
}: IAppLandingPage) => (
  <PageLayout>
    <DetailHeader
      variant="page"
      backLink={false}
      title={app.name}
      id={app.id}
      identity={
        <AppSourceChipComponent repo={app.repo} repoHref={app.repoHref} />
      }
      actions={<Button variant="primary">Create install</Button>}
    />
    <PageContent className="border-t">
      <PageSection className="gap-8">
        <div className="flex flex-col gap-4">
          <SectionHeader
            title="Branches"
            description="Manage app branches and deployment plans."
            actions={<Button variant="secondary">Create branch</Button>}
          />
          <BranchCardsComponent cards={branches} layout="stacked" />
        </div>

        <div className="flex flex-col gap-4">
          <SectionHeader
            title="Installs"
            description="View install configuration and status across this app."
          />
          <AppInstallsList installs={installs} />
        </div>
      </PageSection>
    </PageContent>
  </PageLayout>
)
