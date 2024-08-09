import { Card, Heading, Text, Link } from '@/components'
import { getAppInstalls } from '@/lib'

export default async function AppInstalls({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const installs = await getAppInstalls({ appId, orgId })

  return (
    <>
      {installs.map((install) => (
        <Link key={install.id} href={`/beta/${orgId}/installs/${install.id}`}>
          {install.name}
        </Link>
      ))}
    </>
  )
}
