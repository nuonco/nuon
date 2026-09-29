import { useContext, useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router'
import { ContextTooltip } from '@/components/common/ContextTooltip'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { BreadcrumbQuickNav } from '@/components/navigation/BreadcrumbQuickNav'
import { useBreadcrumb } from '@/hooks/use-breadcrumb'
import { useInstallNested } from '@/hooks/use-install-path'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useNewInstallIA } from '@/hooks/use-new-install-ia'
import { withBreadcrumbQuickNav } from '@/lib/breadcrumb-quick-nav'
import { withAppInstallBreadcrumbs } from '@/lib/install-path'
import { BranchContext } from '@/providers/branch-provider'
import { InstallContext } from '@/providers/install-provider'
import { OrgContext } from '@/providers/org-provider'
import type { TNavLink } from '@/types'

const Separator = () => <Icon variant="CaretRightIcon" className="muted" />

const BreadcrumbItem = ({
  crumb,
  isLast,
  isLoading,
}: {
  crumb: TNavLink
  isLast: boolean
  isLoading: boolean
}) => {
  if (isLoading) {
    return (
      <Skeleton
        height="17px"
        width={`${crumb?.text?.length * 16 * 0.6}px`}
        maxWidth="200px"
      />
    )
  }

  if (crumb.quickNav) {
    return <BreadcrumbQuickNav label={crumb.text} config={crumb.quickNav} />
  }

  return (
    <Text weight="strong">
      <Link
        href={crumb.path}
        isActive={isLast}
        variant="breadcrumb"
        className="truncate max-w-48 inline-block align-bottom"
      >
        {crumb.text}
      </Link>
    </Text>
  )
}

const GAP = 8
const ELLIPSIS_ITEM_WIDTH = 40

function computeCollapseCount(
  availableWidth: number,
  itemWidths: number[]
): number {
  if (itemWidths.length <= 2) return 0

  const totalWidth = itemWidths.reduce(
    (sum, w, i) => sum + w + (i > 0 ? GAP : 0),
    0
  )
  if (totalWidth <= availableWidth) return 0

  let hideCount = 0
  let currentWidth = totalWidth

  for (let i = 1; i < itemWidths.length - 1; i++) {
    if (currentWidth <= availableWidth) break
    currentWidth -= itemWidths[i] + GAP
    if (hideCount === 0) currentWidth += ELLIPSIS_ITEM_WIDTH + GAP
    hideCount++
  }

  return Math.min(hideCount, itemWidths.length - 2)
}

function useCollapsedCount(
  navRef: React.RefObject<HTMLElement | null>,
  listRef: React.RefObject<HTMLOListElement | null>,
  totalItems: number
) {
  const [collapsedCount, setCollapsedCount] = useState(0)
  const itemWidthsRef = useRef<number[]>([])
  const prevTotalRef = useRef(totalItems)

  if (prevTotalRef.current !== totalItems) {
    prevTotalRef.current = totalItems
    itemWidthsRef.current = []
    setCollapsedCount(0)
  }

  useEffect(() => {
    const nav = navRef.current
    const list = listRef.current
    if (!nav || !list) return

    const ro = new ResizeObserver(() => {
      if (itemWidthsRef.current.length === 0) {
        const items = Array.from(list.children) as HTMLElement[]
        if (items.length === 0) return
        itemWidthsRef.current = items.map(
          (el) => el.getBoundingClientRect().width
        )
      }

      const count = computeCollapseCount(nav.clientWidth, itemWidthsRef.current)
      setCollapsedCount(count)
    })
    ro.observe(nav)
    return () => ro.disconnect()
  }, [navRef, listRef, totalItems])

  return collapsedCount
}

export const BreadcrumbNav = () => {
  const { breadcrumbLinks, isLoading } = useBreadcrumb()
  const navRef = useRef<HTMLElement>(null)
  const listRef = useRef<HTMLOListElement>(null)
  const collapsedCount = useCollapsedCount(
    navRef,
    listRef,
    breadcrumbLinks.length
  )

  const firstCrumb = breadcrumbLinks[0]
  if (!firstCrumb) return null

  const shouldCollapse = collapsedCount > 0
  const collapsedCrumbs = shouldCollapse
    ? breadcrumbLinks.slice(1, 1 + collapsedCount)
    : []
  const visibleTail = shouldCollapse
    ? breadcrumbLinks.slice(1 + collapsedCount)
    : breadcrumbLinks.slice(1)

  return (
    <nav
      ref={navRef}
      aria-label="Breadcrumb"
      className="flex-1 min-w-0 overflow-hidden"
    >
      <ol ref={listRef} className="flex items-center gap-2 w-max">
        <li className="flex items-center gap-2">
          <BreadcrumbItem
            crumb={firstCrumb}
            isLast={breadcrumbLinks.length === 1}
            isLoading={isLoading}
          />
        </li>

        {shouldCollapse && (
          <li className="flex items-center gap-2">
            <Separator />
            <ContextTooltip
              items={collapsedCrumbs.map((crumb, idx) => ({
                id: `${idx}:${crumb.path}`,
                title: <Text weight="strong">{crumb.text}</Text>,
                href: crumb.path,
              }))}
              position="bottom"
            >
              <Text weight="strong" className="cursor-default">
                …
              </Text>
            </ContextTooltip>
          </li>
        )}

        {visibleTail.map((crumb, idx) => (
          <li key={`${idx}:${crumb.path}`} className="flex items-center gap-2">
            <Separator />
            <BreadcrumbItem
              crumb={crumb}
              isLast={idx === visibleTail.length - 1}
              isLoading={isLoading}
            />
          </li>
        ))}
      </ol>
    </nav>
  )
}

export const Breadcrumbs = ({ breadcrumbs }: { breadcrumbs: TNavLink[] }) => {
  const { updateBreadcrumb } = useBreadcrumb()
  const org = useContext(OrgContext)?.org
  const install = useContext(InstallContext)?.install
  const branch = useContext(BranchContext)?.branch
  const nested = useInstallNested()
  const hasNewAppIA = useNewAppIA()
  const hasNewInstallIA = useNewInstallIA()
  const params = useParams()
  const routedBreadcrumbs = withAppInstallBreadcrumbs(breadcrumbs, {
    nested,
    orgId: org?.id ?? params.orgId,
    appId: install?.app_id ?? params.appId,
    appName: install?.app?.name,
    installId: install?.id ?? params.installId,
  })
  const orgId = org?.id ?? params.orgId
  const appId = install?.app_id ?? params.appId
  const routed = withBreadcrumbQuickNav(routedBreadcrumbs, {
    branch: {
      enabled: hasNewAppIA,
      orgId,
      appId: params.appId,
      resourceId: params.branchId,
      name: branch?.name,
    },
    install: {
      enabled: hasNewInstallIA,
      orgId,
      appId,
      resourceId: install?.id ?? params.installId,
      name: install?.name,
    },
  })
  const key = JSON.stringify(routed)

  useEffect(() => {
    updateBreadcrumb(routed)
  }, [key])

  return <></>
}
