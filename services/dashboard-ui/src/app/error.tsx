'use client'

import { Dashboard, Heading, Link, Text } from '@/components'

export default function Error({ error }) {
  console.error('Error occured', error)
  
  return (
    <Dashboard>
      <main className="flex h-full gap-6 py-6 md:py-12 lg:py-24">
        <div className="flex flex-col gap-6 lg:max-w-2xl">
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
