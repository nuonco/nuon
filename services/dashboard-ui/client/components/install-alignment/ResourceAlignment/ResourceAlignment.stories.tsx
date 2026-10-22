export default {
  title: 'Features / Installs / Alignment / Resource alignment',
}

import {
  alignmentResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import { DriftResourceCard } from '../DriftResourceCard'
import { ResourceAlignment } from './ResourceAlignment'

const sandbox = {
  id: 'sbx-1',
  kind: 'sandbox' as const,
  name: 'sandbox',
  component_id: undefined,
  component_type: undefined,
}

export const Drifted = () => {
  const resource = unappliedResource({
    ...sandbox,
    reason: 'Sandbox configuration changes haven’t been applied.',
  })
  return (
    <div className="p-4">
      <ResourceAlignment
        label="sandbox"
        resource={resource}
        card={
          <DriftResourceCard
            resource={resource}
            branchRunHref="#branch-run"
            onViewPlan={() => {}}
          />
        }
      />
    </div>
  )
}

export const InSync = () => (
  <div className="p-4">
    <ResourceAlignment label="sandbox" resource={alignmentResource(sandbox)} />
  </div>
)

export const RefreshFailed = () => (
  <div className="p-4">
    <ResourceAlignment
      label="sandbox"
      resource={alignmentResource(sandbox)}
      error={{
        error: 'Internal Server Error',
        description: 'Overview unavailable',
        user_error: false,
      }}
    />
  </div>
)

export const Unknown = () => (
  <div className="p-4">
    <ResourceAlignment
      label="stack"
      resource={alignmentResource({
        ...sandbox,
        kind: 'stack',
        category: 'unknown',
      })}
    />
  </div>
)

export const Loading = () => (
  <div className="p-4">
    <ResourceAlignment label="sandbox" loading />
  </div>
)

export const LoadFailed = () => (
  <div className="p-4">
    <ResourceAlignment
      label="sandbox"
      error={{
        error: 'Internal Server Error',
        description: 'unable to get install overview',
        user_error: false,
      }}
      onRetry={() => {}}
    />
  </div>
)
