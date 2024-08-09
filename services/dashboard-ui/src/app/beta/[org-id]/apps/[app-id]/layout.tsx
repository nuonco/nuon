import { Heading, Text, Link } from '@/components'
import { getApp, getOrg } from '@/lib'

export default async function AppLayout({ children, params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, app] = await Promise.all([
    getOrg({ orgId }),
    getApp({ appId, orgId }),
  ])

  return (
    <>
      <header className="flex items-center justify-between">
        <Heading>
          {org.name} / Apps / {app.name}
        </Heading>

        <nav className="flex items-center gap-6">
          <Link href={`/beta/${orgId}/apps/${appId}`}>Configs</Link>
          <Link href={`/beta/${orgId}/apps/${appId}/components`}>
            Components
          </Link>
          <Link href={`/beta/${orgId}/apps/${appId}/installs`}>Installs</Link>
        </nav>
      </header>
      <section>{children}</section>
    </>
  )
}
