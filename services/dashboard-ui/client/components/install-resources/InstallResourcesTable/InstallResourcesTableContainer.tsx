import { useMemo, useState } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponents, getInstallResources } from '@/lib'
import {
  groupComponentResources,
  groupSandboxResources,
  InstallResourcesTable,
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

  const [searchParams] = useSearchParams()
  const [kind, setKind] = useState(searchParams.get('kind') ?? '')
  const [namespace, setNamespace] = useState(searchParams.get('namespace') ?? '')
  const [health, setHealth] = useState(searchParams.get('health') ?? '')

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

  const filteredResources = useMemo(
    () =>
      allResources.filter((r) => {
        if (kind && r?.kind !== kind) return false
        if (namespace && r?.namespace !== namespace) return false
        if (health && r?.health !== health) return false
        return true
      }),
    [allResources, kind, namespace, health]
  )

  const componentGroups = useMemo(
    () => groupComponentResources(filteredResources, componentNames),
    [filteredResources, componentNames]
  )
  const sandboxGroups = useMemo(
    () => groupSandboxResources(filteredResources),
    [filteredResources]
  )

  return (
    <InstallResourcesTable
      componentGroups={componentGroups}
      sandboxGroups={sandboxGroups}
      isLoading={isLoading}
      kind={kind}
      namespace={namespace}
      health={health}
      kindOptions={kindOptions}
      namespaceOptions={namespaceOptions}
      onKindChange={setKind}
      onNamespaceChange={setNamespace}
      onHealthChange={setHealth}
    />
  )
}
