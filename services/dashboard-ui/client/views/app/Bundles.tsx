import { BundlesTable } from '@/components/apps/bundles/BundlesTable'
import { CreateBundleButton } from '@/components/apps/bundles/CreateBundle'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Bundles = () => {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <PageSection>
      <PageTitle title={`Releases | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/releases`, text: 'Releases' },
        ]}
      />

      <div className="flex items-start justify-between">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Releases
          </Text>
          <Text variant="subtext" theme="neutral">
            Publish immutable application versions, then package them for
            customer-managed installs.
          </Text>
        </HeadingGroup>
        <CreateBundleButton />
      </div>

      <BundlesTable />
    </PageSection>
  )
}
