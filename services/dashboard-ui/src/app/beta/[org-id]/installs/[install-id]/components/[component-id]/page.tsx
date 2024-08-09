import { Card, Heading, Text, Link } from '@/components'
import { getInstallComponent } from '@/lib'

export default async function InstallComponent({ params }) {
  const installComponentId = params?.['component-id'] as string
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const installComponent = await getInstallComponent({
    installComponentId,
    installId,
    orgId,
  })

  return (
    <>
      <Card>
        <Heading>Deploy history</Heading>
        <div>
          {installComponent?.install_deploys?.map((deploy) => (
            <Link
              key={deploy.id}
              href={`/beta/${orgId}/installs/${installId}/components/${installComponentId}/deploys/${deploy.id}`}
            >
              {deploy.id}
            </Link>
          ))}
        </div>
      </Card>
    </>
  )
}
