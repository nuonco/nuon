import classNames from 'classnames'
import React, { type FC, Suspense } from 'react'
import { getAPIVersion } from '@/lib'
import { VERSION } from '@/utils'
import { Text } from '@/components/Typography'

export const NuonVersions: FC<React.HTMLAttributes<HTMLDivElement>> = async ({
  className,
  ...props
}) => {
  const apiVersion = await getAPIVersion().catch((error) => {
    console.error(error)
    return {
      git_ref: 'unknown',
      version: 'unknown',
    }
  })

  return (
    <div
      {...props}
      className={classNames('flex gap-2', {
        [`${className}`]: Boolean(className),
      })}
    >
      <Suspense fallback={<Text variant="med-8">Loading Nuon version...</Text>}>
        <>
          <Text variant="reg-12">API: {apiVersion.version}</Text>
          <Text variant="reg-12">UI: {VERSION}</Text>
        </>
      </Suspense>
    </div>
  )
}
