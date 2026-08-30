import { Card } from '@/components/common/Card'
import { CompositeError } from '@/components/common/CompositeError/CompositeError'
import {
  KeyValueList,
  KeyValueListSkeleton,
} from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
import type { TAppConfig, TInstallStackVersionWithCompositeError } from '@/types'

export interface IGenerateStackDetails {
  appConfig?: TAppConfig
  isLoading: boolean
  stackVersion?: TInstallStackVersionWithCompositeError
}

export const GenerateStackDetails = ({
  appConfig,
  isLoading,
  stackVersion,
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

  return (
    <div className="flex flex-col gap-6">
      {stackVersion?.composite_error ? (
        <CompositeError error={stackVersion.composite_error} />
      ) : null}
      <Card>
        <Text>Stack template details</Text>
        {isLoading ? (
          <KeyValueListSkeleton />
        ) : (
          <KeyValueList values={values} />
        )}
      </Card>
    </div>
  )
}
