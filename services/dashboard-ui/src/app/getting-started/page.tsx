import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { Code, DashboardHeader, Heading, Link, Text } from '@/components'

export default withPageAuthRequired(
  async function GettingStarted() {
    return (
      <div className="h-100vh p-6 max-w-6xl m-auto">
        <DashboardHeader />
        <div className="px-4 pt-12 w-full max-w-md flex flex-col gap-4">
          <Heading variant="title">Started using Nuon today</Heading>
          <Text>
            To get started using Nuon download the{' '}
            <Link href="https://docs.nuon.co/cli" target="_blank">
              Nuon CLI
            </Link>{' '}
            login using the CLI and create your Nuon org using the{' '}
            <Code variant="inline">nuon orgs create</Code> command.
          </Text>
        </div>
      </div>
    )
  },
  { returnTo: '/' }
)
