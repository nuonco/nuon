import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { InstallDeployRow, type IInstallDeployRow } from './InstallDeployRow'
import { StepStatePlaceholder } from '../../shared/StepStatePlaceholder'
import { StepBlock, StepRowList } from '../../shared/StepLayout'

export interface IDeployGroupStep {
  groupName: string
  totalInstalls: number
  deployedCount: number
  rows: IInstallDeployRow[]
  emptyMessage?: string
  variant?: 'group' | 'preview'
}

export const DeployGroupStep = ({
  groupName,
  totalInstalls,
  deployedCount,
  rows,
  emptyMessage,
  variant = 'group',
}: IDeployGroupStep) => {
  const title =
    variant === 'preview' ? (
      <span className="font-semibold text-cool-grey-900 dark:text-white">
        Preview install
      </span>
    ) : (
      <>
        install group:{' '}
        <span className="font-semibold text-cool-grey-900 dark:text-white">
          {groupName}
        </span>
      </>
    )

  return (
    <>
      <StepBlock>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Icon
              variant="PackageIcon"
              size={16}
              className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0"
            />
            <Text variant="body" theme="neutral">
              {title}
            </Text>
            <Text variant="subtext" theme="neutral">
              {totalInstalls} {totalInstalls === 1 ? 'install' : 'installs'}
            </Text>
          </div>
          {totalInstalls > 0 && (
            <Text variant="subtext" family="mono" theme="neutral">
              {deployedCount} / {totalInstalls} deployed
            </Text>
          )}
        </div>

        {rows.length === 0 && emptyMessage ? (
          <StepStatePlaceholder variant="loading">
            {emptyMessage}
          </StepStatePlaceholder>
        ) : null}
      </StepBlock>

      {rows.length > 0 && (
        <StepRowList>
          {rows.map((row) => (
            <InstallDeployRow key={row.installId} {...row} />
          ))}
        </StepRowList>
      )}
    </>
  )
}
