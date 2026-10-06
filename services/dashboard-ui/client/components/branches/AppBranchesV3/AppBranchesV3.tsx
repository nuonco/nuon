import { useEffect, type ReactNode } from 'react'
import { useLocation, useNavigate } from 'react-router'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { BranchDetailActionsComponent } from '@/components/branches/BranchDetailActions'
import { BranchHeaderMeta } from '@/components/branches/BranchHeaderMeta'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { PageContent } from '@/components/layout/PageContent'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { SubNav } from '@/components/navigation/SubNav'
import type { TNavItem } from '@/types'
import { latestRolloutFixture } from './fixtures'
import { RolloutPage } from './RolloutPage'

const BASE_PATH = '/org-1/apps/app-1/branches/brn-1'

const PAGE_TITLES: Record<string, string> = {
  '': 'Overview',
  installs: 'Installs',
  rollout: 'Rollout',
  runs: 'Previous runs',
  settings: 'Settings',
}

const pageKey = (pathname: string) =>
  pathname
    .replace(/\/{2,}/g, '/')
    .slice(BASE_PATH.length)
    .replace(/^\/|\/$/g, '')
    .split('/')[0]

const PagePlaceholder = () => {
  const { pathname } = useLocation()
  const key = pageKey(pathname)
  const title = PAGE_TITLES[key]
  if (key === 'rollout') return <RolloutPage rollout={latestRolloutFixture} />
  if (!title) return null

  return (
    <PageSection>
      <SectionHeader title={title} />
      <Text variant="subtext" theme="neutral">
        {title} page placeholder
      </Text>
    </PageSection>
  )
}

const useBranchBasePath = () => {
  const { pathname } = useLocation()
  const navigate = useNavigate()
  useEffect(() => {
    if (!pathname.startsWith(BASE_PATH)) {
      navigate(BASE_PATH, { replace: true })
    }
  }, [pathname, navigate])
}

const NAV_LINKS: TNavItem[] = [
  { path: '/', iconVariant: 'GraphIcon', text: 'Overview' },
  { path: '/installs', iconVariant: 'CubeIcon', text: 'Installs' },
  { path: '/rollout', iconVariant: 'StackIcon', text: 'Rollout' },
  { path: '/runs', iconVariant: 'ListIcon', text: 'Previous runs' },
  { path: '/settings', iconVariant: 'GearIcon', text: 'Settings' },
  {
    type: 'section',
    label: 'App template',
    defaultOpen: true,
    collapsible: false,
  },
  { path: '/inputs', iconVariant: 'ListChecksIcon', text: 'Inputs' },
  { path: '/components', iconVariant: 'CardsIcon', text: 'Components' },
  { path: '/actions', iconVariant: 'TerminalWindowIcon', text: 'Actions' },
  { path: '/runbooks', iconVariant: 'BookIcon', text: 'Runbooks' },
  { path: '/sandbox', iconVariant: 'ShippingContainerIcon', text: 'Sandboxes' },
  { path: '/policies', iconVariant: 'ShieldCheckIcon', text: 'Policies' },
  { path: '/roles', iconVariant: 'FileLockIcon', text: 'Roles' },
  { path: '/labels', iconVariant: 'TagIcon', text: 'Labels' },
  { path: '/readme', iconVariant: 'BookOpenIcon', text: 'README' },
]

export interface IAppBranchesV3 {
  children?: ReactNode
}

export const AppBranchesV3 = ({ children }: IAppBranchesV3) => {
  useBranchBasePath()

  return (
    <PageLayout>
      <DetailHeader
        variant="page"
        backLink={false}
        title="acme-platform"
        identity={
          <BranchHeaderMeta
            configuration={
              <Text
                as="span"
                variant="subtext"
                weight="strong"
                flex
                nowrap
                className="gap-1"
              >
                main
                <Icon variant="CaretUpDownIcon" size={12} />
              </Text>
            }
            repo="acme/platform"
            gitBranch="main"
            directory="/deploy"
            trigger="Every push"
          />
        }
        actions={
          <BranchDetailActionsComponent
            isTriggerPending={false}
            onTriggerRun={() => {}}
            onTriggerPreviewModal={() => {}}
          />
        }
      />
      <PageContent className="border-t" variant="row">
        <SubNav
          basePath={BASE_PATH}
          links={NAV_LINKS}
          storageKey="subnav:app-branches-v3"
          pinLastGroup
        />
        <div className="flex min-w-0 flex-1 flex-col">
          {children ?? <PagePlaceholder />}
        </div>
      </PageContent>
    </PageLayout>
  )
}
