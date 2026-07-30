import { PageSection } from '@/components/layout/PageSection'
import { PageTitle } from '@/components/navigation/PageTitle'
import { InstallBranchesSection } from '@/components/installs/InstallBranches'
import { useInstall } from '@/hooks/use-install'

export const InstallBranches = () => {
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle title={`App branches | ${install?.name}`} />
      <InstallBranchesSection install={install} />
    </PageSection>
  )
}
