import { Dashboard, Heading, Link, Text } from '@/components'

export default function NotFound() {
  return (
    <Dashboard>
      <main className="flex h-full gap-6">
        <div className="flex flex-col gap-6 p-0 md:p-12 lg:p-24 lg:max-w-2xl">
          <Heading variant="title" level={1}>
            Nuon organization not found
          </Heading>
          <Text className="text-lg leading-relaxed">
            There was an issue loading your Nuon organization.
            <br />
            If this issue persist please contact Nuon{' '}
            <Link href="mailto:team@nuon.co">support@nuon.co</Link>
          </Text>
          <Link className="text-sm" href="/">
            Return to homepage
          </Link>
        </div>
      </main>
    </Dashboard>
  )
}
