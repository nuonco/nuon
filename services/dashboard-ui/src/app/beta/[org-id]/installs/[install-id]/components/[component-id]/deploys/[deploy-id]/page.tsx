import { Card, Heading, Text, Link } from '@/components'
import { getOrg, getInstall, getDeploy } from '@/lib'

export default async function InstallComponentDeploy({ params }) {
  const deployId = params?.['deploy-id'] as string
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const deploy = await getDeploy({ deployId, installId, orgId })

  return <>
    <Card>
      <Heading>Deploy details</Heading>
      <Text>
        {deploy.install_deploy_type}
      </Text>
    </Card>
  </>
}
