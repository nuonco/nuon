import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Sandbox = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/sandbox`,
            text: 'Sandbox',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install sandbox
        </Text>
      </HeadingGroup>

      {/* install sandbox content here */}
    </PageSection>
  )
}
