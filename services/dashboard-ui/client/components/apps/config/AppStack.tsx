import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { PropertyGrid } from "@/components/common/PropertyGrid"
import { Text } from '@/components/common/Text'
import type { TAppConfig } from '@/types'

export interface IAppStack {
  appConfig: TAppConfig
}

export const AppStack = ({ appConfig }: IAppStack) => {
  const stackConfig = appConfig?.stack

  return (
    <div className="flex flex-col gap-6">
      <div className="flex gap-6 items-start justify-start">
        <LabeledValue label="Type">
          <Text family="mono" variant="subtext">
            {stackConfig?.type}
          </Text>
        </LabeledValue>
        <LabeledValue label="Name">
          <Text variant="subtext">{stackConfig?.name}</Text>
        </LabeledValue>
      </div>

      {stackConfig?.runner_nested_template_url ? (
        <LabeledValue label="Runner template URL">
          <Text variant="subtext">
            <Link href={stackConfig?.runner_nested_template_url} isExternal>
              {stackConfig?.runner_nested_template_url}
            </Link>
          </Text>
        </LabeledValue>
      ) : null}

      {stackConfig?.vpc_nested_template_url ? (
        <LabeledValue label="VPC template URL">
          <Text variant="subtext">
            <Link href={stackConfig?.vpc_nested_template_url} isExternal>
              {stackConfig?.vpc_nested_template_url}
            </Link>
          </Text>
        </LabeledValue>
      ) : null}

    {stackConfig?.custom_nested_stacks?.length ? (
        <PropertyGrid
          values={stackConfig?.custom_nested_stacks?.map((s) => ({
            name: s?.name,
            template_url: s?.template_url,
            contents_hash: s?.contents_hash,
          }))}
          gridTemplate="1fr 2fr 2fr"
        />
      ) : null}
    </div>
  )
}
