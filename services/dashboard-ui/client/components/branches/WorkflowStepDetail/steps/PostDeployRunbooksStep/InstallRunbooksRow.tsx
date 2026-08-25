import { CloudPlatform } from '@/components/common/CloudPlatform'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TCloudPlatform, TInstall } from '@/types'

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
  const platform = (install?.cloud_platform?.toLowerCase() as TCloudPlatform | undefined) || 'unknown'

  return (
    <div className="flex flex-col">
      <div className="flex items-center gap-3 px-4 py-3">
        {installHref ? (
          <Text variant="body" weight="strong" nowrap className="truncate">
            <Link href={installHref}>{install?.name || installId}</Link>
          </Text>
        ) : (
          <Text variant="body" weight="strong" nowrap className="block truncate">
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
      </div>

      <div className="flex flex-col border-t bg-cool-grey-50 dark:bg-dark-grey-900">
        {runbooks.map((runbook, idx) => (
          <div
            key={runbook.runbookId || runbook.runbookName}
            className="flex items-center gap-3 pl-6 pr-4 py-2.5"
          >
            <Text variant="subtext" family="mono" theme="neutral" className="shrink-0">
              {idx + 1}
            </Text>
            <Text variant="body" nowrap className="truncate">
              {runbook.runbookName}
            </Text>

            <div className="flex-1" />

            <Status status={runbook.status || 'pending'} variant="badge" className="shrink-0" />

            {runbook.workflowHref && (
              <Text variant="subtext" className="shrink-0">
                <Link href={runbook.workflowHref}>View run</Link>
              </Text>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
