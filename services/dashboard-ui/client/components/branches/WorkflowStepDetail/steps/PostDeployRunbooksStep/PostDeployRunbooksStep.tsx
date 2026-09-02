import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import {
  InstallRunbooksRow,
  type IInstallRunbooksRow,
} from './InstallRunbooksRow'
import { StepStatePlaceholder } from '../../shared/StepStatePlaceholder'
import { StepBlock, StepRowList } from '../../shared/StepLayout'

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
    <>
      <StepBlock>
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

        {rows.length === 0 ? (
          emptyMessage ? (
            <StepStatePlaceholder variant="loading">
              {emptyMessage}
            </StepStatePlaceholder>
          ) : statusDescription ? (
            <Text variant="subtext" theme="neutral">
              {statusDescription}
            </Text>
          ) : (
            <StepStatePlaceholder variant="pending">
              No post-deploy runbooks ran for this group.
            </StepStatePlaceholder>
          )
        ) : null}
      </StepBlock>

      {rows.length > 0 && (
        <StepRowList>
          {rows.map((row) => (
            <InstallRunbooksRow key={row.installId} {...row} />
          ))}
        </StepRowList>
      )}
    </>
  )
}
