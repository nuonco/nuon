'use client'
import { useEffect, type FC } from 'react'
import type { FallbackProps } from 'react-error-boundary'
import { Warning } from '@phosphor-icons/react'
import { Link } from '@/components/old/Link'
import { Text } from '@/components/old/Typography'

export const ErrorFallback: FC<FallbackProps> = ({ error }) => {
  useEffect(() => {
    console.error('Error occured: ', error)
  }, [error])

  return (
    <div className="flex flex-col gap-2 lg:max-w-xl">
      <Text className="text-red-800 dark:text-red-500 !gap-2" variant="med-14">
        <Warning size={18} />
        An error occurred
      </Text>
      <Text variant="reg-14">
        {error?.message || 'An unknown error occured.'}
      </Text>
      <Text variant="reg-14">
        If this issue persist please contact us at{' '}
        <Link href="mailto:team@nuon.co">support@nuon.co</Link>
      </Text>

      <Link className="text-base" href="/">
        Return to homepage
      </Link>
    </div>
  )
}
