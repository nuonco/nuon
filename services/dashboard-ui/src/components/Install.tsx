'use client'

import React, { type FC } from 'react'
import {
  Card,
  Heading,
  InstallStatus,
  InstallPlatformType,
  InstallRegion,
  Link,
  Text,
} from '@/components'
import { useInstallContext } from '@/context'

export const InstallCard: FC = () => {
  const {
    install: { id, org_id, ...install },
  } = useInstallContext()

  return (
    <Card>
      <div className="flex flex-col gap-1 flex-auto">
        <InstallStatus isCompact />
        <Text className="text-gray-500" variant="overline">
          {id}
        </Text>
        <Heading>{install?.name}</Heading>
      </div>
      <div className="flex flex-col gap-1">
        <Text variant="caption">
          <b>App:</b> {install?.app?.name}
        </Text>

        <Text variant="caption">
          <b>Platform:</b> <InstallPlatformType />
        </Text>

        <Text variant="caption">
          <b>Region:</b> <InstallRegion />
        </Text>
      </div>

      <Text variant="caption">
        <Link href={`/dashboard/${org_id}/${id}`}>Details</Link>
      </Text>
    </Card>
  )
}
