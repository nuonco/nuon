import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { InstallRunbooksRow, type IInstallRunbooksRow } from './InstallRunbooksRow'
import { StepStatePlaceholder } from '../../shared/StepStatePlaceholder'

export interface IPostDeployRunbooksStep {
  groupName: string
  runbookNames: string[]
  rows: IInstallRunbooksRow[]
  emptyMessage?: string
  statusDescription?: string
}

export const PostDeployRunbooksStep = ({
  groupName,
  runbookNames,
  rows,
  emptyMessage,
  statusDescription,
}: IPostDeployRunbooksStep) => {
  const completed = rows.filter((row) =>
    row.runbooks.every((runbook) => runbook.status === 'success')
  ).length

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Icon
            variant="PlayIcon"
            size={16}
            className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0"
          />
          <Text variant="body" theme="neutral">
            post-deploy runbooks:{' '}
            <span className="font-semibold text-cool-grey-900 dark:text-white">
              {runbookNames.join(' → ')}
            </span>
          </Text>
          <Text variant="subtext" theme="neutral">
            on {groupName}
          </Text>
        </div>
        {rows.length > 0 && (
          <Text variant="subtext" family="mono" theme="neutral">
            {completed} / {rows.length} installs
          </Text>
        )}
      </div>

      {rows.length > 0 ? (
        <div className="border rounded-[10px] divide-y overflow-hidden">
          {rows.map((row) => (
            <InstallRunbooksRow key={row.installId} {...row} />
          ))}
        </div>
      ) : emptyMessage ? (
        <StepStatePlaceholder variant="loading">{emptyMessage}</StepStatePlaceholder>
      ) : statusDescription ? (
        <div className="p-3 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
          <Text variant="base">{statusDescription}</Text>
        </div>
      ) : (
        <StepStatePlaceholder variant="pending">
          No post-deploy runbooks ran for this group.
        </StepStatePlaceholder>
      )}
    </div>
  )
}
