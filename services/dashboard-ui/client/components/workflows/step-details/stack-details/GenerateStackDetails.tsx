'use client'

import { useQuery } from '@tanstack/react-query'
import { Card } from '@/components/common/Card'
import {
  KeyValueList,
  KeyValueListSkeleton,
} from '@/components/common/KeyValueList'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig } from '@/lib'
import type { TAppConfig } from '@/types'

export const GenerateStackDetails = () => {
  const { install } = useInstall()
  const { org } = useOrg()
  const { data: appConfig = {} as TAppConfig, isLoading } = useQuery<TAppConfig>({
    queryKey: ['app-config', org?.id, install?.app_id, install?.app_config_id],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!install?.app_config_id,
  })

  const values = [
    { key: 'name', value: appConfig.stack?.name },
    { key: 'description', value: appConfig.stack?.description },
    {
      key: 'runner_nested_template_url',
      value: appConfig.stack?.runner_nested_template_url,
    },
    {
      key: 'vpc_nested_template_url',
      value: appConfig.stack?.vpc_nested_template_url,
    },
    { key: 'type', value: appConfig.stack?.type },
  ]

  return isLoading ? (
    <GenerateStackDetailsSkeleton />
  ) : (
    <Card>
      <Text>Stack template details</Text>
      <KeyValueList values={values} />
    </Card>
  )
}

export const GenerateStackDetailsSkeleton = () => {
  return (
    <Card>
      <Skeleton height="24px" width="142px" />
      <KeyValueListSkeleton />
    </Card>
  )
}
