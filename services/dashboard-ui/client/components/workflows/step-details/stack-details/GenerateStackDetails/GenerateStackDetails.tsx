import { Card } from '@/components/common/Card'
import {
  KeyValueList,
  KeyValueListSkeleton,
} from '@/components/common/KeyValueList'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TAppConfig } from '@/types'

export interface IGenerateStackDetails {
  appConfig?: TAppConfig
  isLoading: boolean
}

export const GenerateStackDetails = ({
  appConfig,
  isLoading,
}: IGenerateStackDetails) => {
  const values = [
    { key: 'name', value: appConfig?.stack?.name },
    { key: 'description', value: appConfig?.stack?.description },
    {
      key: 'runner_nested_template_url',
      value: appConfig?.stack?.runner_nested_template_url,
    },
    {
      key: 'vpc_nested_template_url',
      value: appConfig?.stack?.vpc_nested_template_url,
    },
    { key: 'type', value: appConfig?.stack?.type },
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
