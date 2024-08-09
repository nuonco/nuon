import { Card, Heading, Text, Link } from '@/components'
import { getSandboxRun } from '@/lib'

export default async function SandboxRuns({ params }) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const runId = params?.['run-id'] as string
  const sandboxRun = await getSandboxRun({ installId, orgId, runId })

  return (
    <>
      <Card>
        <Heading>Sandbox run details</Heading>
        <Text>{sandboxRun.run_type}</Text>
      </Card>
    </>
  )
}
