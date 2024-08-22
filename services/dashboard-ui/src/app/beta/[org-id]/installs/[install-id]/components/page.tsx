import { Card, Heading, Text, Link } from '@/components'
import { getInstall } from '@/lib'

export default async function InstallComponents({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const install = await getInstall({ orgId, installId })

  return (
    <>
      {install.install_components?.length &&
        install.install_components.map((component) => (
          <Link
            key={component.id}
            href={`/beta/${orgId}/installs/${installId}/components/${component.id}`}
          >
            {component.component?.name}
          </Link>
        ))}
    </>
  )
}
