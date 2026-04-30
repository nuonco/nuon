import { OCIArtifactDetails } from '@/components/deploys/OCIArtifactDetails'
import { useDeploy } from '@/hooks/use-deploy'

export const DeployArtifactTab = () => {
  const { deploy } = useDeploy()
  return <OCIArtifactDetails artifact={deploy?.oci_artifact} />
}
