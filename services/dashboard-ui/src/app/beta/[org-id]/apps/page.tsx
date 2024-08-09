import { Card, Heading, Text, Link } from '@/components'
import { getApps, getOrg } from '@/lib'

export default async function Apps({ params }) {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })
  const apps = await getApps({ orgId })

  return (
    <>
      <header>
        <Heading>{org.name} / Apps</Heading>
      </header>
      <section>
        {apps.map((app) => (
          <Link key={app.id} href={`/beta/${orgId}/apps/${app.id}`}>
            {app.name}
          </Link>
        ))}
      </section>
    </>
  )
}
