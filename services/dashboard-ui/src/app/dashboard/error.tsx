'use client'

import { useEffect } from 'react'
import { Heading, Link, Text } from '@/components'

export default function Error({
  error,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <div className="flex flex-col flex-auto items-center justify-start py-24">
      <Heading>Not found</Heading>
      <Text>{error.message}</Text>
      <Text>
        If this issue persist conntact us at{' '}
        <Link href="mailto:support@nuon.co">support@nuon.co</Link>
      </Text>
    </div>
  )
}
