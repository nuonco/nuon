import { OCIArtifactDetails } from '@/components/deploys/OCIArtifactDetails'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'

export const DeployArtifactTab = () => {
  const { deploy } = useDeploy()
  const { install } = useInstall()
  return (
    <>
      <PageTitle segments={['Deploy artifact', install?.name]} />
      <OCIArtifactDetails artifact={deploy?.oci_artifact} />
    </>
  )
}
