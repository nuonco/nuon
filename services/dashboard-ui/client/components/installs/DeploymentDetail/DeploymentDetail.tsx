import type { ReactNode } from 'react'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import {
  DetailPage,
  type TDetailPageVariant,
} from '@/components/layout/DetailPage'
import type { TInstallDeploymentRecord, TNavLink, TWorkflow } from '@/types'
import { humanize } from '@/utils/string-utils'

export const DEPLOYMENT_DETAIL_TABS: TNavLink[] = [
  { path: '/', text: 'Changes' },
  { path: '/template-updates', text: 'Template updates' },
  { path: '/workflow', text: 'Workflow' },
]

interface IDeploymentDetail {
  activeTabIndex?: number
  banners?: ReactNode
  basePath: string
  branchHref?: string
  children: ReactNode
  deployment?: TInstallDeploymentRecord
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
  variant = 'section',
  workflow,
}: IDeploymentDetail) => {
  const title =
    deployment?.title ||
    workflow?.name ||
    humanize(workflow?.type) ||
    'Deployment'
  const branch = deployment?.app_branch

  return (
    <DetailPage
      variant={variant}
      header={
        <DetailHeader
          title={title}
          status={
            <Status
              status={deployment?.status || workflow?.status?.status}
              variant="badge"
            />
          }
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
        />
      }
      banners={banners}
      tabNav={{
        activeIndex: activeTabIndex,
        basePath,
        tabs: DEPLOYMENT_DETAIL_TABS,
      }}
    >
      {children}
    </DetailPage>
  )
}
