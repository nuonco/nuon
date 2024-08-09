import { Card, Heading, Text, Link } from '@/components'
import { getOrg, getInstall } from '@/lib'

export default async function InstallLayout({ children, params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const org = await getOrg({ orgId })
  const install = await getInstall({ orgId, installId })

  return (
    <>
      <header className="flex items-center justify-between">
        <Heading>
          {org.name} / Installs / {install.name}
        </Heading>

        <nav className="flex items-center gap-6">
          <Link href={`/beta/${orgId}/installs/${installId}`}>Status</Link>
          <Link href={`/beta/${orgId}/installs/${installId}/components`}>
            Components
          </Link>
        </nav>
      </header>
      <section>{children}</section>
    </>
  )
}
