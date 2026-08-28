import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  groupComponentResources,
  groupSandboxResources,
  healthFacetCounts,
  InstallResourcesTableComponent,
  matchesResourceSearch,
} from '@/components/install-resources/InstallResourcesTable'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import type { TInstallResource } from '@/types'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'

export const CustomerManagedSnapshotResources = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const [searchParams, setSearchParams] = useSearchParams()
  const data = snapshot?.snapshot
  const observedAt = data?.health?.observed_at
  const kind = searchParams.get('kind') ?? ''
  const namespace = searchParams.get('namespace') ?? ''
  const health = searchParams.get('health') ?? ''
  const search = searchParams.get('q') ?? ''

  const resources = useMemo<TInstallResource[]>(() => {
    const componentResources = (data?.health?.components ?? []).flatMap(
      (component) =>
        (component.resources ?? []).map((resource) => ({
          ...resource,
          install_component_id:
            component.install_component_id ??
            component.component_id ??
            component.component_name,
          owner_name: component.component_name,
          observed_at: observedAt,
          source: 'component',
        }))
    )
    const sandboxResources = (data?.health?.sandbox_releases ?? []).flatMap(
      (release) =>
        (release.resources ?? []).map((resource) => ({
          ...resource,
          owner_name: release.release_name,
          observed_at: observedAt,
          source: 'sandbox',
        }))
    )
    return [...componentResources, ...sandboxResources]
  }, [data?.health, observedAt])

  const componentNames = useMemo(
    () =>
      Object.fromEntries(
        (data?.health?.components ?? []).map((component) => [
          component.install_component_id ??
            component.component_id ??
            component.component_name ??
            'unknown',
          component.component_name ?? 'Unknown component',
        ])
      ),
    [data?.health?.components]
  )
  const filteredResources = resources.filter((resource) => {
    if (kind && resource.kind !== kind) return false
    if (namespace && resource.namespace !== namespace) return false
    if (health && resource.health !== health) return false
    return matchesResourceSearch(resource, search)
  })
  const setFilter = (key: string) => (value: string) => {
    setSearchParams(
      (params) => {
        if (value) params.set(key, value)
        else params.delete(key)
        return params
      },
      { replace: true }
    )
  }

  return (
    <CustomerManagedSnapshotContent>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Resources
        </Text>
        <Text variant="subtext" theme="neutral">
          Resources and health captured by the customer runner in this support
          snapshot.
        </Text>
      </HeadingGroup>

      <Card>
        <div className="flex flex-wrap items-center justify-between gap-4">
          <span className="flex items-center gap-2">
            <Text weight="strong">Captured health</Text>
            <Status
              variant="badge"
              status={
                !data?.health
                  ? 'unknown'
                  : data.health.cluster_access_error
                    ? 'degraded'
                    : 'active'
              }
            />
          </span>
          {observedAt ? (
            <span className="flex items-center gap-2">
              <Text variant="subtext" theme="neutral">
                Observed
              </Text>
              <Time time={observedAt} format="relative" variant="subtext" />
            </span>
          ) : null}
        </div>
        {data?.health?.cluster_access_error ? (
          <Text variant="subtext" theme="error">
            {data.health.cluster_access_error}
          </Text>
        ) : null}
      </Card>

      <InstallResourcesTableComponent
        componentGroups={groupComponentResources(
          filteredResources,
          componentNames
        )}
        sandboxGroups={groupSandboxResources(filteredResources)}
        healthCounts={healthFacetCounts(filteredResources)}
        isLoading={false}
        kind={kind}
        namespace={namespace}
        health={health}
        search={search}
        kindOptions={
          Array.from(
            new Set(resources.map((resource) => resource.kind).filter(Boolean))
          ).sort() as string[]
        }
        namespaceOptions={
          Array.from(
            new Set(
              resources.map((resource) => resource.namespace).filter(Boolean)
            )
          ).sort() as string[]
        }
        onKindChange={setFilter('kind')}
        onNamespaceChange={setFilter('namespace')}
        onHealthChange={setFilter('health')}
      />
    </CustomerManagedSnapshotContent>
  )
}
