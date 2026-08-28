import { CloudPlatform } from '@/components/common/CloudPlatform'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TCloudPlatform, TInstall } from '@/types'
import { StepRow } from '../../shared/StepLayout'

export interface IRunbookRun {
  runbookId?: string
  runbookName: string
  status?: string
  workflowHref?: string
}

export interface IInstallRunbooksRow {
  installId: string
  install?: TInstall
  installHref?: string
  runbooks: IRunbookRun[]
}

export const InstallRunbooksRow = ({
  installId,
  install,
  installHref,
  runbooks,
}: IInstallRunbooksRow) => {
  const platform =
    (install?.cloud_platform?.toLowerCase() as TCloudPlatform | undefined) ||
    'unknown'

  return (
    <div className="flex flex-col">
      <StepRow>
        {installHref ? (
          <Text variant="body" weight="strong" nowrap className="truncate">
            <Link href={installHref} variant="inline">{install?.name || installId}</Link>
          </Text>
        ) : (
          <Text
            variant="body"
            weight="strong"
            nowrap
            className="block truncate"
          >
            {install?.name || installId}
          </Text>
        )}

        {platform !== 'unknown' && (
          <CloudPlatform
            platform={platform}
            colorVariant="color"
            displayVariant="icon-only"
            iconSize="20"
            className="shrink-0"
          />
        )}
      </StepRow>

      <div className="flex flex-col border-t bg-cool-grey-50 dark:bg-dark-grey-900">
        {runbooks.map((runbook, idx) => (
          <div
            key={runbook.runbookId || runbook.runbookName}
            className="flex items-center gap-3 py-2.5 pl-8 pr-4 sm:pl-10 sm:pr-6"
          >
            <Text
              variant="subtext"
              family="mono"
              theme="neutral"
              className="shrink-0"
            >
              {idx + 1}
            </Text>
            <Text variant="body" nowrap className="truncate">
              {runbook.runbookName}
            </Text>

            <div className="flex-1" />

            <Status
              status={runbook.status || 'pending'}
              variant="badge"
              className="shrink-0"
            />

            {runbook.workflowHref && (
              <Link href={runbook.workflowHref} className="shrink-0">View run</Link>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
