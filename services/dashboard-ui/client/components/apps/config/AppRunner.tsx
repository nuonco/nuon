import { Badge } from '@/components/common/Badge'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { KeyValueList } from '@/components/common/KeyValueList'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TAppConfig } from '@/types'
import { objectToKeyValueArray } from '@/utils/data-utils'

// Mirrors DefaultAWSPhoneHomeScript in
// services/ctl-api/internal/app/installs/worker/activities/get_phonehome_script.go —
// the API only returns phone_home_script_url when an app-level override is set.
const DEFAULT_AWS_PHONE_HOME_SCRIPT_URL =
  'https://raw.githubusercontent.com/nuonco/runner/refs/tags/aws-v0.1.4/scripts/aws/phonehome.py'

export interface IAppRunner {
  appConfig: TAppConfig
}

export const AppRunner = ({ appConfig }: IAppRunner) => {
  const runnerConfig = appConfig?.runner
  const runnerEnvVars = objectToKeyValueArray(runnerConfig?.env_vars)

  const phoneHomeOverride = runnerConfig?.phone_home_script_url
  const phoneHomeUrl =
    phoneHomeOverride ||
    (runnerConfig?.cloud_platform === 'aws'
      ? DEFAULT_AWS_PHONE_HOME_SCRIPT_URL
      : undefined)

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
            <Link href={runnerConfig?.init_script} isExternal>
              View script
            </Link>
          </LabeledValue>
        ) : null}

        {phoneHomeUrl ? (
          <LabeledValue
            label={
              <span className="flex items-center gap-2">
                <Text variant="subtext" theme="neutral">
                  Phone home script
                </Text>
                {phoneHomeOverride ? (
                  <Badge size="sm" theme="info">
                    override
                  </Badge>
                ) : (
                  <Text variant="subtext" theme="neutral">
                    default
                  </Text>
                )}
              </span>
            }
          >
            <Link href={phoneHomeUrl} isExternal>
              View script
            </Link>
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
