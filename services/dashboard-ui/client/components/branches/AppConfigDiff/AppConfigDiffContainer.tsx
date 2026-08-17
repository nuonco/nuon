import { useQuery } from '@tanstack/react-query'
import { useContext, useEffect, useRef, useState } from 'react'
import type { TConfigDiffFocus } from '@/components/approvals/plan-diffs/config-diff-focus'
import { useOrg } from '@/hooks/use-org'
import { AppContext } from '@/providers/app-provider'
import { scrollElementIntoView } from '@/utils/scroll'
import { getAppConfigs, getAppConfigDiff } from '@/lib'
import { extractSections, computeSummary } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { AppConfigDiffCard } from './AppConfigDiffCard'

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
      <AppConfigDiffCard
        sections={sections}
        summary={summary}
        versionLabel={versionLabel}
        isLoading={isLoading && !diffData}
        isOpen={cardOpen}
        focus={focus}
        className={className}
      />
    </div>
  )
}
