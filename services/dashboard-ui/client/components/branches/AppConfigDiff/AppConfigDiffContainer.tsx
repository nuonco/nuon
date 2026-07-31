import { useQuery } from '@tanstack/react-query'
import { useContext, useEffect, useRef, useState } from 'react'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import type { TConfigDiffFocus } from '@/components/approvals/plan-diffs/config-diff-focus'
import { useOrg } from '@/hooks/use-org'
import { AppContext } from '@/providers/app-provider'
import { cn } from '@/utils/classnames'
import { scrollElementIntoView } from '@/utils/scroll'
import { getAppConfigs, getAppConfigDiff } from '@/lib'
import { AppConfigDiff, extractSections, computeSummary } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'

interface IAppConfigDiffContainer {
  appConfigId: string
  oldConfigId?: string
  appId?: string
  className?: string
  focus?: TConfigDiffFocus | null
}

export const AppConfigDiffContainer = ({ appConfigId, oldConfigId: oldConfigIdProp, appId: appIdProp, className, focus }: IAppConfigDiffContainer) => {
  const { org } = useOrg()
  const appCtx = useContext(AppContext)
  const appId = appIdProp || appCtx?.app?.id
  const cardRef = useRef<HTMLDivElement>(null)
  const [cardOpen, setCardOpen] = useState(true)

  const { data: recentConfigs } = useQuery({
    queryKey: ['app-configs', org?.id, appId],
    queryFn: () => getAppConfigs({ orgId: org!.id, appId: appId!, limit: 10 }),
    enabled: !!org?.id && !!appId && !!appConfigId,
  })

  const previousConfigs = (recentConfigs || []).filter((c) => c.id !== appConfigId)
  const oldConfig = oldConfigIdProp
    ? (recentConfigs || []).find((c) => c.id === oldConfigIdProp) ?? previousConfigs[0]
    : previousConfigs[0]
  const oldConfigId = oldConfigIdProp || oldConfig?.id
  const newConfig = (recentConfigs || []).find((c) => c.id === appConfigId)

  const { data: diffData, isLoading } = useQuery({
    queryKey: ['app-config-diff', org?.id, appId, appConfigId, oldConfigId],
    queryFn: () =>
      getAppConfigDiff({
        orgId: org!.id,
        appId: appId!,
        configId: appConfigId,
        oldConfigId,
      }),
    enabled: !!org?.id && !!appId && !!appConfigId,
    retry: 1,
  })

  useEffect(() => {
    if (!focus) return
    setCardOpen(true)
    const timer = setTimeout(() => {
      scrollElementIntoView(cardRef.current, { block: 'start' })
    }, 60)
    return () => clearTimeout(timer)
  }, [focus?.nonce])

  const sections = diffData?.diff ? extractSections(diffData.diff) : []
  const summary = sections.length > 0 ? computeSummary(sections) : (diffData?.summary || null)

  const newVersion = newConfig?.version
  const oldVersion = oldConfig?.version
  const versionLabel =
    newVersion != null
      ? oldVersion != null
        ? `v${oldVersion} → v${newVersion}`
        : `v${newVersion}`
      : null

  return (
    <div ref={cardRef} className="scroll-mt-4">
      <Expand
        id="config-changes"
        isOpen={cardOpen}
        className={cn(
          'border rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden',
          className
        )}
        headerClassName="px-5 py-4"
        heading={
          <div className="flex items-center gap-3 w-full">
            <Text variant="h3" weight="strong">
              Config changes
            </Text>
            {versionLabel && (
              <Text variant="subtext" theme="neutral" family="mono">
                {versionLabel}
              </Text>
            )}
            {summary && (
              <ChangeCountSummary
                added={summary.added}
                updated={summary.changed}
                removed={summary.removed}
                className="ml-auto"
              />
            )}
          </div>
        }
      >
        <div className="p-5 border-t max-h-[70vh] overflow-y-auto">
          <AppConfigDiff
            sections={sections}
            summary={null}
            isLoading={isLoading && !diffData}
            defaultSectionsOpen={false}
            focus={focus}
          />
        </div>
      </Expand>
    </div>
  )
}
