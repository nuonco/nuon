import { Card, Heading, Text, Link } from '@/components'
import { getOrg, getInstall } from '@/lib'

export default async function InstallComponents({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const org = await getOrg({ orgId })
  const install = await getInstall({ orgId, installId })

  return (
    <>
      {install.install_components.map((component) => (
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
