import type { Metadata } from 'next'
import { Suspense } from 'react'
import { notFound } from 'next/navigation'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getAppBranch, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { BranchWorkflowCanvas } from './branch-workflow-canvas'

type TBranchCanvasPageProps = TPageProps<'org-id' | 'app-id' | 'branch-id'>

export async function generateMetadata({
  params,
}: TBranchCanvasPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params
  const { data: branch } = await getAppBranch({ appId, branchId, orgId })

  const branchName = branch?.name || branchId
  return {
    title: `${branchName} - Canvas | Branches | Nuon`,
  }
}

export default async function AppBranchCanvasPage({
  params,
}: TBranchCanvasPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params

  // Try to fetch real data, but fallback gracefully for canvas demo
  const [{ data: branch }, { data: app }, { data: org }] =
    await Promise.all([
      getAppBranch({ appId, branchId, orgId }),
      getApp({ appId, orgId }),
      getOrg({ orgId }),
    ])

  // Use branch name if available, otherwise use the branchId
  const branchName = branch?.name || branchId

  return (
    <PageSection isScrollable className="max-w-full overflow-x-hidden">
      <div className="w-[80%] flex flex-col gap-4 md:gap-6">
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
              path: `/${orgId}/apps/${appId}/branches`,
              text: 'Branches',
            },
            {
              path: `/${orgId}/apps/${appId}/branches/${branchId}`,
              text: branchName,
            },
            {
              path: `/${orgId}/apps/${appId}/branches/${branchId}/canvas`,
              text: 'Canvas',
            },
          ]}
        />

        <HeadingGroup>
          <div className="flex items-center gap-3">
            <Icon variant="GitBranch" size={24} />
            <Text variant="h3" weight="stronger">
              {branchName} - App Branch Run
            </Text>
          </div>
          <Text variant="subtext" theme="neutral">
            Workflow execution for app branch deployment
          </Text>
        </HeadingGroup>

        <div className="w-full overflow-x-hidden">
          <ErrorBoundary fallback={<>Error loading workflow canvas</>}>
            <Suspense fallback={<div>Loading canvas...</div>}>
              <BranchWorkflowCanvas branchId={branchId} />
            </Suspense>
          </ErrorBoundary>
        </div>
      </div>
    </PageSection>
  )
}