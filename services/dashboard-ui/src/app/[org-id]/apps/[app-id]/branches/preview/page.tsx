import { Suspense } from 'react'
import type { Metadata } from 'next'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { Card } from '@/components/common/Card'
import { InstallGroupsConfig } from './install-groups-config'

export const metadata: Metadata = {
  title: 'Branch Install Configuration Preview | Nuon',
}

export default async function BranchInstallConfigPreviewPage({
  params,
}: {
  params: Promise<{ 'org-id': string; 'app-id': string }>
}) {
  const { 'org-id': orgId, 'app-id': appId } = await params

  return (
    <PageSection isScrollable>
      <div className="flex flex-col gap-6 max-w-7xl mx-auto">
        {/* Preview Header */}
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
          <Text variant="base" weight="strong" className="mb-2">
            Install Configuration Preview
          </Text>
          <Text variant="subtext" theme="neutral">
            This is a standalone preview of the install grouping configuration UI.
            Configure deployment groups for your app branch below.
          </Text>
        </div>

        {/* Mock Branch Info */}
        <Card>
          <div className="flex flex-col gap-2">
            <Text variant="h4" weight="strong">
              Branch: production
            </Text>
            <div className="flex gap-4 text-sm text-gray-600 dark:text-gray-400">
              <span>Repository: nuonco/platform</span>
              <span>Branch: main</span>
              <span>Directory: services/api</span>
            </div>
          </div>
        </Card>

        {/* Install Groups Configuration */}
        <Suspense fallback={<div>Loading...</div>}>
          <InstallGroupsConfig orgId={orgId} appId={appId} />
        </Suspense>
      </div>
    </PageSection>
  )
}
