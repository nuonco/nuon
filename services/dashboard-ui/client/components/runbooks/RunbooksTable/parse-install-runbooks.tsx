import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RunRunbookButton } from '@/components/runbooks/RunRunbook'
import type { TInstallRunbook } from '@/lib/ctl-api/installs/runbooks'
import type { TRunbookRow } from './RunbooksTable'

export function parseInstallRunbooksToTableData(
  runbooks: TInstallRunbook[],
  orgId: string,
  installId: string,
  labelColors?: Record<string, string>,
  removed = false
): TRunbookRow[] {
  return runbooks.map((ir) => {
    const basePath = `/${orgId}/installs/${installId}`
    const runbook = ir.runbook
    const runbookId = ir.runbook_id ?? ir.id
    const href = `${basePath}/runbooks/${runbookId}`
    const latestRun = ir.runs?.[0]
    const workflowId =
      latestRun?.install_workflow_id ?? latestRun?.install_workflow?.id
    const latestRunHref = workflowId
      ? `${basePath}/workflows/${workflowId}`
      : null

    return {
      runbookId,
      runbookName: runbook?.name ?? '',
      description: runbook?.description ? (
        <div className="max-w-[250px] line-clamp-2">
          <Text variant="subtext" theme="neutral">
            {runbook.description}
          </Text>
        </div>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      labels:
        runbook?.labels && Object.keys(runbook.labels).length > 0 ? (
          <span className="flex flex-wrap gap-1">
            {Object.keys(runbook.labels)
              .sort()
              .map((k) => (
                <LabelBadge
                  key={k}
                  labelKey={k}
                  labelValue={runbook.labels[k]}
                  size="sm"
                  customColor={labelColors?.[k]}
                />
              ))}
          </span>
        ) : (
          <Icon variant="MinusIcon" />
        ),
      lastUpdated: runbook?.updated_at ? (
        <Text flex nowrap className="gap-2">
          <Icon variant="CalendarBlankIcon" />
          <Time
            time={runbook.updated_at}
            format="relative"
            variant="subtext"
            nowrap
          />
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      lastRun: latestRun ? (
        <Text flex nowrap className="gap-2">
          <Icon variant="CalendarBlankIcon" />
          <Time
            time={latestRun.created_at}
            format="relative"
            variant="subtext"
            nowrap
          />
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      href,
      removed,
      actions: (
        <Dropdown
          alignment="right"
          buttonText=""
          buttonClassName="!p-1"
          icon={<Icon variant="DotsThreeVerticalIcon" />}
          id={runbookId}
          variant="ghost"
        >
          <Menu>
            {removed ? (
              <Button
                isMenuButton
                disabled
                className="w-full"
                tooltipProps={{
                  className: 'block !w-full',
                  position: 'left',
                  tipContent:
                    "This runbook is no longer in the install's app config version.",
                }}
              >
                Run runbook
                <Icon variant="PlayIcon" />
              </Button>
            ) : (
              <RunRunbookButton installRunbook={ir} isMenuButton>
                Run runbook
              </RunRunbookButton>
            )}
            {latestRunHref ? (
              <Button href={latestRunHref} isMenuButton>
                Latest run
              </Button>
            ) : null}
            <Button href={href} isMenuButton>
              View details
            </Button>
          </Menu>
        </Dropdown>
      ),
    }
  })
}
