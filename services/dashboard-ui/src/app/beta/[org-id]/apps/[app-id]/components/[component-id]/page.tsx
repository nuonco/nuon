import { Heading, Link } from '@/components'
import { getComponentBuilds } from '@/lib'

export default async function AppComponent({ params }) {
  const appId = params?.['app-id'] as string
  const componentId = params?.['component-id'] as string
  const orgId = params?.['org-id'] as string
  const builds = await getComponentBuilds({ componentId, orgId })

  return (
    <section className="px-6 py-8">
      <div>
        <Heading>Build history</Heading>
        <div>
          {builds.map((build) => (
            <Link
              key={build.id}
              href={`/beta/${orgId}/apps/${appId}/components/${componentId}/builds/${build.id}`}
            >
              {build.id}
            </Link>
          ))}
        </div>
      </div>
    </section>
  )
}
