import type { ReactNode } from 'react'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import {
  DetailPage,
  type TDetailPageVariant,
} from '@/components/layout/DetailPage'
import type { TInstallDeploymentRecord, TNavLink, TWorkflow } from '@/types'
import { humanize } from '@/utils/string-utils'
import { DeploymentRunStatus } from './DeploymentProgress'
import {
  DEPLOYMENT_TABS,
  deploymentSteps,
  deploymentTabOrder,
  type TDeploymentTab,
} from './deployment-progress'

export const DEPLOYMENT_DETAIL_TABS: TNavLink[] = deploymentTabOrder().map(
  (key) => DEPLOYMENT_TABS[key]
)

interface IDeploymentDetail {
  activeTabIndex?: number
  banners?: ReactNode
  basePath: string
  branchHref?: string
  children: ReactNode
  deployment?: TInstallDeploymentRecord
  search?: string
  tabOrder?: TDeploymentTab[]
  variant?: TDetailPageVariant
  workflow: TWorkflow
}

export const DeploymentDetail = ({
  activeTabIndex,
  banners,
  basePath,
  branchHref,
  children,
  deployment,
  search = '',
  tabOrder,
  variant = 'section',
  workflow,
}: IDeploymentDetail) => {
  const title =
    deployment?.title ||
    workflow?.name ||
    humanize(workflow?.type) ||
    'Deployment'
  const branch = deployment?.app_branch
  const run = {
    status: workflow?.status?.status ?? deployment?.status ?? 'unknown',
    activity:
      workflow?.status?.status_human_description ?? deployment?.summary ?? '',
    steps: deploymentSteps(workflow),
  }

  return (
    <DetailPage
      variant={variant}
      className="[&>.tab-nav]:gap-2 md:[&>.tab-nav]:gap-6 [&>.tab-nav>a]:px-1 md:[&>.tab-nav>a]:px-3"
      header={
        <DetailHeader
          backLink={false}
          title={title}
          description={run.activity}
          status={<DeploymentRunStatus run={run} />}
          id={workflow?.id}
          identity={
            <Time
              time={deployment?.created_at || workflow?.created_at}
              format="relative"
              variant="subtext"
              theme="info"
            />
          }
          metadata={
            <>
              <LabeledValue label="Type">
                {humanize(deployment?.type || workflow?.type)}
              </LabeledValue>
              {branch && branchHref ? (
                <LabeledValue label="App branch">
                  <Link href={branchHref}>{branch.name}</Link>
                </LabeledValue>
              ) : null}
            </>
          }
        >
          <Link
            href={`${basePath.replace(/\/deployments\/[^/]+$/, '/deployments')}${search}`}
          >
            Back to deployments
          </Link>
        </DetailHeader>
      }
      banners={banners}
      tabNav={{
        activeIndex: activeTabIndex,
        basePath,
        tabs: (tabOrder ?? deploymentTabOrder(run.status)).map((key) => ({
          ...DEPLOYMENT_TABS[key],
          path: search
            ? `${DEPLOYMENT_TABS[key].path === '/' ? '' : DEPLOYMENT_TABS[key].path}${search}`
            : DEPLOYMENT_TABS[key].path,
        })),
      }}
    >
      {children}
    </DetailPage>
  )
}
