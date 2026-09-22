import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const NewInstallPlaceholderBody = ({ title }: { title: string }) => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={[title, install?.name]} />
      <Text theme="neutral">Placeholder content.</Text>
    </>
  )
}

export const NewInstallPlaceholder = ({
  path,
  title,
}: {
  path: string
  title: string
}) => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle segments={[title, install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}${path}`,
            text: title,
          },
        ]}
      />
      <SectionHeader title={title} />
      <Text theme="neutral">Placeholder content.</Text>
    </PageSection>
  )
}
