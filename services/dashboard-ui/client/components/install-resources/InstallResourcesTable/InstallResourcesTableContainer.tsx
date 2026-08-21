import { useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponents, getInstallResources } from '@/lib'
import {
  groupComponentResources,
  groupSandboxResources,
  healthFacetCounts,
  InstallResourcesTable,
  matchesHealthFilter,
  matchesResourceSearch,
} from './InstallResourcesTable'

export const InstallResourcesTableContainer = ({
  pollInterval = 15000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const { install } = useInstall()

  // The URL is the source of truth for filters, both ways: deep links (like a
  // degraded-health message linking to ?health=degraded) apply even when this
  // page is already mounted, and picking a filter updates the URL so it's
  // shareable and survives back/forward.
  const [searchParams, setSearchParams] = useSearchParams()
  const kind = searchParams.get('kind') ?? ''
  const namespace = searchParams.get('namespace') ?? ''
  const health = searchParams.get('health') ?? ''
  const search = searchParams.get('q') ?? ''

  const setFilter = useCallback(
    (key: string) => (value: string) => {
      setSearchParams(
        (params) => {
          if (value) params.set(key, value)
          else params.delete(key)
          return params
        },
        { replace: true }
      )
    },
    [setSearchParams]
  )

  const { data: resources, isLoading } = useQuery({
    queryKey: ['install-resources', org?.id, install?.id],
    queryFn: () =>
      getInstallResources({ orgId: org!.id, installId: install!.id }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const { data: componentsResult } = useQuery({
    queryKey: ['install-components-for-resources', org?.id, install?.id],
    queryFn: () =>
      getInstallComponents({
        orgId: org!.id,
        installId: install!.id,
        limit: 100,
        offset: 0,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const componentNames = useMemo(() => {
    const map: Record<string, string> = {}
    componentsResult?.data?.forEach((component) => {
      if (component?.id) {
        map[component.id] = component?.component?.name || component.id
      }
    })
    return map
  }, [componentsResult])

  // Structured twin of the "(downstream of X)" description suffix — present
  // only while both the component and its dependency are bad.
  const downstreamOf = useMemo(() => {
    const map: Record<string, string> = {}
    componentsResult?.data?.forEach((component) => {
      const value = (component?.health_status_v2?.metadata as Record<string, unknown> | undefined)
        ?.downstream_of
      if (component?.id && typeof value === 'string' && value) {
        map[component.id] = value
      }
    })
    return map
  }, [componentsResult])

  const allResources = resources ?? []

  const kindOptions = useMemo(
    () =>
      Array.from(
        new Set(allResources.map((r) => r?.kind).filter((v): v is string => !!v))
      ).sort(),
    [allResources]
  )
  const namespaceOptions = useMemo(
    () =>
      Array.from(
        new Set(
          allResources.map((r) => r?.namespace).filter((v): v is string => !!v)
        )
      ).sort(),
    [allResources]
  )

  // The chips are facet counts for the health axis, so they stay stable while
  // you toggle between them — only the other axes narrow them.
  const scopedResources = useMemo(
    () =>
      allResources.filter((r) => {
        if (kind && r?.kind !== kind) return false
        if (namespace && r?.namespace !== namespace) return false
        return matchesResourceSearch(r, search)
      }),
    [allResources, kind, namespace, search]
  )

  const healthCounts = useMemo(
    () => healthFacetCounts(scopedResources),
    [scopedResources]
  )

  const filteredResources = useMemo(
    () => scopedResources.filter((r) => matchesHealthFilter(r, health)),
    [scopedResources, health]
  )

  const componentGroups = useMemo(
    () => groupComponentResources(filteredResources, componentNames, downstreamOf),
    [filteredResources, componentNames, downstreamOf]
  )
  const sandboxGroups = useMemo(
    () => groupSandboxResources(filteredResources),
    [filteredResources]
  )

  return (
    <InstallResourcesTable
      componentGroups={componentGroups}
      sandboxGroups={sandboxGroups}
      healthCounts={healthCounts}
      isLoading={isLoading}
      kind={kind}
      namespace={namespace}
      health={health}
      search={search}
      kindOptions={kindOptions}
      namespaceOptions={namespaceOptions}
      onKindChange={setFilter('kind')}
      onNamespaceChange={setFilter('namespace')}
      onHealthChange={setFilter('health')}
    />
  )
}
