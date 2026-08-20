import { KeyValueList } from '@/components/common/KeyValueList'
import { Link } from '@/components/common/Link'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Text } from '@/components/common/Text'
import type { TAppConfig, TAppStackConfig, TKeyValue } from '@/types'

export interface IAppStack {
  appConfig: TAppConfig
}

const scalarKeys = [
  'type',
  'name',
  'description',
  'vpc_nested_template_url',
  'runner_nested_template_url',
] as const satisfies readonly (keyof TAppStackConfig)[]

// deployment_scope is omitted from the response whenever it is the default, and
// is deliberately never normalized on write so that configs stored before the
// field existed do not show a spurious diff on the next sync. Rendering the
// absent case as an em dash would leave the scope unreadable for exactly the
// configs where it is most often unstated.
const defaultDeploymentScope = 'resource_group'

const stackKeyValues = (stackConfig: TAppStackConfig): TKeyValue[] => {
  const values: TKeyValue[] = scalarKeys.map((key) => ({
    key,
    value: stackConfig?.[key] ?? '',
  }))

  if (stackConfig?.type === 'azure-bicep') {
    values.push({
      key: 'deployment_scope',
      value: stackConfig?.deployment_scope || defaultDeploymentScope,
    })
  }

  return values
}

export const AppStack = ({ appConfig }: IAppStack) => {
  const stackConfig = appConfig?.stack

  if (!stackConfig) {
    return null
  }

  return (
    <div className="flex flex-col gap-6">
      <KeyValueList values={stackKeyValues(stackConfig)} />

      {stackConfig?.custom_nested_stacks?.length ? (
        <div className="flex flex-col gap-2">
          <Text variant="subtext" weight="strong">
            Custom nested stacks
          </Text>
          <PropertyGrid
            values={[...stackConfig.custom_nested_stacks]
              .sort((a, b) => (a.index ?? 0) - (b.index ?? 0))
              .map((s) => ({
                index: s?.index,
                name: s?.name,
                template_url: s?.template_url,
                contents_hash: s?.contents_hash,
              }))}
            columns={[
              { key: 'index', header: 'Index' },
              { key: 'name', header: 'Name' },
              {
                key: 'template_url',
                header: 'Template URL',
                render: (value) =>
                  value ? (
                    <Text variant="subtext">
                      <Link href={String(value)} isExternal>
                        {String(value)}
                      </Link>
                    </Text>
                  ) : null,
              },
              { key: 'contents_hash', header: 'Contents hash' },
            ]}
            gridTemplate="min-content 1fr 2fr 2fr"
          />
        </div>
      ) : null}
    </div>
  )
}
