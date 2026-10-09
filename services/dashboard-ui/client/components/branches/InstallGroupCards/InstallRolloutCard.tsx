import { Button } from '@/components/common/Button'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { installStatusBucket } from './install-status'
import { CommitRange, type ICommitRange } from './CommitRange'
import type { TTrackInstall } from '@/components/branches/BranchOverview/RolloutTrack'

const CARD_CLASS =
  'flex w-full items-center justify-between gap-3 rounded-md border bg-elevation-2 px-4 py-3 text-left shadow-sm'

const InstallCardFace = ({
  install,
  commit,
}: {
  install: TTrackInstall
  commit?: ICommitRange
}) => (
  <>
    <span className="flex min-w-0 items-baseline gap-2">
      <Text variant="body" weight="strong" className="truncate">
        {install.name}
      </Text>
      {commit ? <CommitRange commit={commit} /> : null}
    </span>
    <span className="flex shrink-0 items-center gap-2">
      <Status status={install.status} />
    </span>
  </>
)

export const InstallRolloutCard = ({
  install,
  commit,
  onSelect,
}: {
  install: TTrackInstall
  commit?: ICommitRange
  onSelect?: (install: TTrackInstall) => void
}) => {
  if (installStatusBucket(install.status) === 'pending') {
    return (
      <Tooltip
        className="!w-full"
        position="top"
        tipContent="This rollout has not started yet."
      >
        <div className={`${CARD_CLASS} cursor-default`}>
          <InstallCardFace install={install} commit={commit} />
        </div>
      </Tooltip>
    )
  }

  if (!onSelect || !install.workflowId) {
    return (
      <div className={CARD_CLASS}>
        <InstallCardFace install={install} commit={commit} />
      </div>
    )
  }

  return (
    <Button
      variant="ghost"
      className="!h-auto !w-full !whitespace-normal !justify-between !gap-3 !px-4 !py-3 text-left !rounded-md !shadow-sm !bg-elevation-2 !text-inherit !border !border-[color:var(--border-color)] !transition-[background-color] !duration-fast !ease-cubic hover:!bg-elevation-3"
      onClick={() => onSelect(install)}
    >
      <InstallCardFace install={install} commit={commit} />
    </Button>
  )
}
