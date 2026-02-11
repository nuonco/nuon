import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { getAppConfig } from '@/lib'
import type { TInstall } from '@/types'

export const StackConfig = async ({
  install,
  orgId,
}: {
  install: TInstall
  orgId: string
}) => {
  const { data: config, error } = await getAppConfig({
    appId: install?.app_id,
    appConfigId: install?.app_config_id,
    orgId,
    recurse: true,
  })

  return !config && error ? (
    <StackConfigError />
  ) : (
    <Card>
      <Text weight="strong">Current stack config</Text>

      <div className="grid grid-cols-6 gap-3">
        <LabeledValue label="App config version">
          {config?.version?.toString()}
        </LabeledValue>

        <LabeledValue label="Stack type">{config?.stack?.type}</LabeledValue>

        <LabeledValue label="Stack name">{config?.stack?.name}</LabeledValue>

        {config?.stack?.runner_nested_template_url ? (
          <LabeledValue
            className="col-span-6"
            label="Runner nested template URL"
          >
            <Text variant="subtext">
              <Link href={config?.stack?.runner_nested_template_url} isExternal>
                {config?.stack?.runner_nested_template_url}
              </Link>
            </Text>
          </LabeledValue>
        ) : null}

        {config?.stack?.vpc_nested_template_url ? (
          <LabeledValue className="col-span-6" label="VPC nested template URL">
            <Text variant="subtext">
              <Link href={config?.stack?.vpc_nested_template_url} isExternal>
                {config?.stack?.vpc_nested_template_url}
              </Link>
            </Text>
          </LabeledValue>
        ) : null}
      </div>
    </Card>
  )
}

export const StackConfigSkeleton = () => (
  <Card>
    <Skeleton width="135px" height="24px" />

    <div className="grid grid-cols-6 gap-3">
      <LabeledValue label={<Skeleton height="17px" width="100px" />}>
        <Skeleton height="17px" width="20px" />
      </LabeledValue>
      <LabeledValue label={<Skeleton height="17px" width="60px" />}>
        <Skeleton height="17px" width="110px" />
      </LabeledValue>
      <LabeledValue label={<Skeleton height="17px" width="65px" />}>
        <Skeleton height="17px" width="180px" />
      </LabeledValue>

      <LabeledValue
        className="col-span-6"
        label={<Skeleton height="17px" width="156px" />}
      >
        <Skeleton height="17px" width="627px" />
      </LabeledValue>

      <LabeledValue
        className="col-span-6"
        label={<Skeleton height="17px" width="140px" />}
      >
        <Skeleton height="17px" width="840px" />
      </LabeledValue>
    </div>
  </Card>
)

export const StackConfigError = ({
  message = 'Unable to load this install stack config',
  title = 'No stack config',
}: {
  message?: string
  title?: string
}) => (
  <Card>
    <EmptyState emptyMessage={message} emptyTitle={title} variant="table" />
  </Card>
)
