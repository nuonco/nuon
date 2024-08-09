import { GoArrowRight } from 'react-icons/go'
import { Dashboard, Heading, Text, Link } from '@/components'
import { getOrg, getInstalls } from '@/lib'

export default async function Installs({ params }) {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })
  const installs = await getInstalls({ orgId })

  return (
    <>
      <header>
        <Heading>{org.name} / Installs</Heading>
      </header>
      <section>
        {installs.map((install) => (
          <Link key={install.id} href={`/beta/${orgId}/installs/${install.id}`}>
            {install.name}
          </Link>
        ))}
      </section>
    </>
  )
}
