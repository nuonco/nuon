import type { Metadata } from 'next'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getInstall, getOrg } from '@/lib'
import { TPageProps } from '@/types'
import {
  StackConfig,
  StackConfigError,
  StackConfigSkeleton,
} from './stack-config'
import { InstallStacksTable, InstallStacksTableSkeleton } from './stacks-table'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install }: any = await getInstall({ installId, orgId })

  return {
    title: `Stacks | ${install.name} | Nuon`,
  }
}

export default async function InstallStack({ params }: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const [{ data: install }, { data: org }] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({
      orgId,
    }),
  ])

  const containerId = 'stack-page'
  return (
    <PageSection id={containerId} className="!pb-24" isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/stacks`,
            text: 'Stacks',
          },
        ]}
      />

      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install stacks
        </Text>
        <Text variant="subtext" theme="neutral">
          View your install stack config and versions below.
        </Text>
      </HeadingGroup>

      <AsyncBoundary
        loadingFallback={<StackConfigSkeleton />}
        errorFallback={<StackConfigError />}
      >
        <StackConfig orgId={orgId} install={install} />
      </AsyncBoundary>

      <div className="flex flex-col gap-4">
        <Text weight="strong">Install stack versions</Text>
        <AsyncBoundary
          loadingFallback={<InstallStacksTableSkeleton />}
          errorFallback={
            <span className="text-md">Unable to load install stacks</span>
          }
        >
          <InstallStacksTable installId={install?.id} orgId={orgId} />
        </AsyncBoundary>
      </div>
      <BackToTop containerId={containerId} />
    </PageSection>
  )
}
