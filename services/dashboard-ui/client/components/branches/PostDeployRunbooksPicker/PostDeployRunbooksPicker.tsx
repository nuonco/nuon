import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TRunbook } from '@/lib/ctl-api/apps/runbooks/get-runbooks'

export interface IPostDeployRunbooksPicker {
  runbooks: TRunbook[]
  loadingRunbooks: boolean
  selectedRunbookIds: string[]
  onChange: (runbookIds: string[]) => void
  disabled?: boolean
}

export const PostDeployRunbooksPicker = ({
  runbooks,
  loadingRunbooks,
  selectedRunbookIds,
  onChange,
  disabled,
}: IPostDeployRunbooksPicker) => {
  const runbooksById = new Map(runbooks.map((runbook) => [runbook.id, runbook]))
  const available = runbooks.filter(
    (runbook) => !selectedRunbookIds.includes(runbook.id)
  )

  const add = (runbookId: string) => {
    if (!runbookId || selectedRunbookIds.includes(runbookId)) return
    onChange([...selectedRunbookIds, runbookId])
  }

  const remove = (idx: number) => {
    onChange(selectedRunbookIds.filter((_, i) => i !== idx))
  }

  const move = (idx: number, delta: number) => {
    const target = idx + delta
    if (target < 0 || target >= selectedRunbookIds.length) return
    const next = [...selectedRunbookIds]
    ;[next[idx], next[target]] = [next[target], next[idx]]
    onChange(next)
  }

  if (loadingRunbooks) {
    return (
      <div className="flex flex-col gap-1">
        <Label htmlFor="post-deploy-runbooks">
          <Text variant="body" className="font-medium">
            Post-deploy runbooks
          </Text>
        </Label>
        <Skeleton height="36px" />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <Label htmlFor="post-deploy-runbooks">
        <Text variant="body" className="font-medium">
          Post-deploy runbooks (optional)
        </Text>
      </Label>
      <Text variant="subtext" theme="neutral">
        Run after each install deploys, in order. A failure fails the install and
        stops the rollout.
      </Text>

      {runbooks.length === 0 ? (
        <Banner theme="warn">No runbooks found for this app.</Banner>
      ) : (
        <>
          {selectedRunbookIds.length > 0 && (
            <ol className="flex flex-col gap-2">
              {selectedRunbookIds.map((runbookId, idx) => {
                const runbook = runbooksById.get(runbookId)
                return (
                  <li
                    key={runbookId}
                    className="flex items-center gap-3 px-3 py-2.5 rounded-md bg-cool-grey-50 dark:bg-dark-grey-900"
                  >
                    <Text
                      variant="subtext"
                      theme="neutral"
                      className="shrink-0 font-mono"
                    >
                      {idx + 1}
                    </Text>
                    <Text
                      variant="body"
                      weight="strong"
                      nowrap
                      className="truncate flex-1"
                    >
                      {runbook?.name || runbookId}
                    </Text>
                    <Button
                      variant="ghost"
                      size="xs"
                      onClick={() => move(idx, -1)}
                      disabled={disabled || idx === 0}
                      title="Move earlier"
                      className="!p-1 shrink-0"
                    >
                      <Icon variant="ArrowUpIcon" size={14} />
                    </Button>
                    <Button
                      variant="ghost"
                      size="xs"
                      onClick={() => move(idx, 1)}
                      disabled={disabled || idx === selectedRunbookIds.length - 1}
                      title="Move later"
                      className="!p-1 shrink-0"
                    >
                      <Icon variant="ArrowDownIcon" size={14} />
                    </Button>
                    <Button
                      variant="ghost"
                      size="xs"
                      onClick={() => remove(idx)}
                      disabled={disabled}
                      title={`Remove ${runbook?.name || runbookId}`}
                      className="!p-1 shrink-0"
                    >
                      <Icon variant="XIcon" size={14} />
                    </Button>
                  </li>
                )
              })}
            </ol>
          )}

          {available.length > 0 && (
            <Select
              id="post-deploy-runbooks"
              value=""
              onChange={(e) => add(e.target.value)}
              disabled={disabled}
              options={[
                { value: '', label: 'Add a runbook…' },
                ...available.map((runbook) => ({
                  value: runbook.id,
                  label: runbook.name,
                })),
              ]}
              searchable
            />
          )}
        </>
      )}
    </div>
  )
}
