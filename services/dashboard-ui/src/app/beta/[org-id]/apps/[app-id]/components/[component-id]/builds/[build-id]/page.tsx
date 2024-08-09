import { Card, Heading, Text, Time } from '@/components'
import { getBuild } from '@/lib'

export default async function AppComponent({ params }) {
  const appId = params?.['app-id'] as string
  const buildId = params?.['build-id'] as string
  const componentId = params?.['component-id'] as string
  const orgId = params?.['org-id'] as string
  const build = await getBuild({ buildId, orgId })

  return (
    <>
      <Card>
        <Heading>Build details</Heading>
        <Text>{build.component_config_version}</Text>
        <Time time={build.created_at} />
        <Time time={build.updated_at} />
      </Card>
    </>
  )
}
