import { CloudPlatform } from '@/components/common/CloudPlatform'
import { KeyValueList } from '@/components/common/KeyValueList'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TAppConfig } from '@/types'
import { objectToKeyValueArray } from '@/utils/data-utils'

export interface IAppRunner {
  appConfig: TAppConfig
}

export const AppRunner = ({ appConfig }: IAppRunner) => {
  const runnerConfig = appConfig?.runner
  const runnerEnvVars = objectToKeyValueArray(runnerConfig?.env_vars)

  return (
    <div className="flex flex-col">
      <div className="flex gap-6 items-start justify-start">
        <LabeledValue label="Platform">
          <CloudPlatform
            variant="subtext"
            platform={runnerConfig?.cloud_platform}
          />
        </LabeledValue>

        <LabeledValue label="Runner type">
          <Text family="mono" variant="subtext">
            {runnerConfig?.app_runner_type}
          </Text>
        </LabeledValue>

        {runnerConfig?.helm_driver ? (
          <LabeledValue label="Helm driver">
            <Text family="mono" variant="subtext">
              {runnerConfig?.helm_driver}
            </Text>
          </LabeledValue>
        ) : null}

        {runnerConfig?.init_script ? (
          <LabeledValue label="Init script">
            <Text variant="subtext">
              <Link href={runnerConfig?.init_script} isExternal>
                View script
              </Link>
            </Text>
          </LabeledValue>
        ) : null}
      </div>

      {runnerEnvVars?.length ? (
        <div>
          <Text variant="subtext" weight="strong">
            Environment variables
          </Text>

          <KeyValueList
            emptyStateProps={{
              emptyTitle: 'No runner env vars',
              emptyMessage:
                'No environment variables configured for this runner',
            }}
            values={runnerEnvVars}
          />
        </div>
      ) : null}
    </div>
  )
}
