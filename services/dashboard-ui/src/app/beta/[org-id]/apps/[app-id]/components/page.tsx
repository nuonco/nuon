import { Heading, Text, Link } from '@/components'
import { getAppComponents } from '@/lib'

export default async function AppComponents({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const components = await getAppComponents({ appId, orgId })

  return (
    <>
      {components.map((component) => (
        <Link
          key={component.id}
          href={`/beta/${orgId}/apps/${appId}/components/${component.id}`}
        >
          {component.name}
        </Link>
      ))}
    </>
  )
}
