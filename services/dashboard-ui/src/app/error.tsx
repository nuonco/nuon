'use client'

import { Dashboard, Heading, Link, Text } from '@/components'

export default function Error({ error }) {
  return (
    <Dashboard>
      <main className="flex h-full gap-6">
        <div className="flex flex-col gap-6 p-0 md:p-12 lg:p-24 lg:max-w-2xl">
          <Heading variant="title" level={1}>
            An error occurred
          </Heading>
          <div>
            <Text className="text-lg leading-relaxed">
              {error?.message || 'An unknown error occured.'}
            </Text>
            <Text className="text-lg leading-relaxed ">
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
