import { Dashboard, EmptyStateGraphic, Link, Text } from '@/components'

export default function NotFound() {
  return (
    <Dashboard>
      <main className="flex h-full gap-6">
        <div className="flex flex-col gap-6 py-6 md:py-12 lg:py-24 lg:max-w-2xl">
          <div>
            <EmptyStateGraphic />
          </div>
          <Text variant="semi-18" level={1}>
            Nuon data not found
          </Text>
          <div>
          <Text className="!text-lg !leading-relaxed">
            There was an issue loading this data.
          </Text>
          <Text className="!text-lg !leading-relaxed">
            If this issue persist please contact Nuon{' '}
            <Link href="mailto:team@nuon.co">support@nuon.co</Link>
          </Text>
          </div>
          <Link className="text-base" href="/">
            Return to homepage
          </Link>
        </div>
      </main>
    </Dashboard>
  )
}
