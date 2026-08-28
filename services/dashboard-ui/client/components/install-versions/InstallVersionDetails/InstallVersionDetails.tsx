import { AppConfigDiff } from '@/components/branches/AppConfigDiff'
import { Badge } from '@/components/common/Badge'
import { Divider } from '@/components/common/Divider'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TInstallAppConfigVersion } from '@/types'
import {
  installVersionSource,
  resolveInstallVersionStatus,
} from './install-version-utils'

export interface IInstallVersionDetails extends IPanel {
  version: TInstallAppConfigVersion
  orgId?: string
  installId?: string
  appId?: string
}

export const InstallVersionDetails = ({
  version,
  orgId,
  installId,
  appId,
  ...props
}: IInstallVersionDetails) => {
  const branchRun = version?.app_branch_run
  const commit = branchRun?.vcs_connection_commit
  const metadata = Object.entries(version?.metadata ?? {})

  return (
    <Panel heading="App config version" size="half" {...props}>
      <div className="flex items-center gap-2">
        <Status variant="badge" status={resolveInstallVersionStatus(version)} />
        <Badge size="sm" theme="neutral">
          {installVersionSource(version)}
        </Badge>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <LabeledValue label="Applied">
          {version?.created_at ? (
            <Time
              variant="subtext"
              time={version.created_at}
              format="relative"
            />
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
        <LabeledValue label="Version">
          <ID>{version?.id}</ID>
        </LabeledValue>
        <LabeledValue label="Old config">
          {version?.old_app_config_id ? (
            <ID>{version.old_app_config_id}</ID>
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
        <LabeledValue label="New config">
          {version?.new_app_config_id ? (
            <ID>{version.new_app_config_id}</ID>
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
      </div>

      {commit ? (
        <LabeledValue label="Commit">
          <span className="flex flex-wrap items-center gap-2">
            {branchRun?.app_branch?.name ? (
              <Badge size="sm" theme="info">
                {branchRun.app_branch.name}
              </Badge>
            ) : null}
            {commit.sha ? (
              <Badge size="sm" variant="code">
                {commit.sha.slice(0, 7)}
              </Badge>
            ) : null}
            <Text variant="subtext" theme="neutral">
              {commit.message}
            </Text>
            {branchRun?.pr_number ? (
              <Badge size="sm" theme="neutral">
                PR #{branchRun.pr_number}
              </Badge>
            ) : null}
          </span>
        </LabeledValue>
      ) : null}

      {metadata.length > 0 ? (
        <LabeledValue label="Metadata">
          <span className="flex flex-wrap items-center gap-2">
            {metadata.map(([key, value]) => (
              <Badge key={key} size="sm" theme="neutral">
                {key}: {value}
              </Badge>
            ))}
          </span>
        </LabeledValue>
      ) : null}

      <div className="flex flex-wrap items-center gap-4">
        {orgId && appId && branchRun?.workflow_id && branchRun?.app_branch?.id ? (
          <Link
            href={`/${orgId}/apps/${appId}/branches/${branchRun.app_branch.id}/runs/${branchRun.workflow_id}`}
          >
            View branch run
          </Link>
        ) : null}
        {orgId && installId && version?.workflow_id ? (
          <Link href={`/${orgId}/installs/${installId}/workflows/${version.workflow_id}`}>
            View workflow
          </Link>
        ) : null}
      </div>

      {version?.new_app_config_id && appId ? (
        <>
          <Divider dividerWord="Config diff" />
          <AppConfigDiff
            appConfigId={version.new_app_config_id}
            oldConfigId={version.old_app_config_id}
            appId={appId}
          />
        </>
      ) : null}
    </Panel>
  )
}
