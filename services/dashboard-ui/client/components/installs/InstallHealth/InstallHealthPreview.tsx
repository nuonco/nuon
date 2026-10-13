import { useEffect, useState } from 'react'
import { InstallHealth } from './InstallHealth'
import { InstallHealthResources } from './InstallHealthResources'
import {
  healthPreviewFixture,
  type THealthPreviewState,
} from './health-preview-fixtures'

export const InstallHealthPreview = ({
  state,
}: {
  state: THealthPreviewState
}) => {
  const [snapshot, setSnapshot] = useState(() => healthPreviewFixture(state))
  useEffect(() => {
    const refreshSnapshot = () => setSnapshot(healthPreviewFixture(state))
    refreshSnapshot()
    const interval = setInterval(refreshSnapshot, 15_000)
    return () => clearInterval(interval)
  }, [state])

  const clusterAccessError =
    state === 'access-error'
      ? 'The runner cannot list Pods in the acme namespace. Restore cluster access to receive a new snapshot.'
      : undefined
  const unavailable = [
    'access-error',
    'no-observations',
    'empty',
    'loading',
  ].includes(state)
  return (
    <InstallHealth
      clusterAccessError={clusterAccessError}
      resources={
        <InstallHealthResources
          resources={
            unavailable
              ? []
              : [
                  ...snapshot.groups.flatMap((group) =>
                    group.items.map((item) => item.resource)
                  ),
                  ...snapshot.sandboxResources,
                  ...snapshot.checks,
                ]
          }
          componentNames={Object.fromEntries(
            snapshot.groups.map((group) => [`instcmp-${group.id}`, group.name])
          )}
          defaultComponent="component:instcmp-api"
          isLoading={state === 'loading'}
          clusterAccessError={clusterAccessError}
          caption="Ladle preview · Illustrative observations · Refresh ~15s"
        />
      }
    />
  )
}
