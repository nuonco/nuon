export default {
  title: 'Branches/WorkflowStepDetail/BuildStep',
}

import { BuildStep } from './BuildStep'

export const Default = () => (
  <BuildStep
    status="success"
    metadata={{
      builds: [
        { component_id: 'c1', component_name: 'api', status: 'success', cache_status: 'cache hit', duration: 2.4, image_digest: 'sha256:9f8e7d6c5b4a3210' },
        { component_id: 'c2', component_name: 'web', status: 'success', cache_status: 'no cache', duration: 41.7, image_digest: 'sha256:1122334455667788' },
        { component_id: 'c3', component_name: 'migrations', status: 'skipped', cache_status: 'cache hit', duration: 0.1 },
      ],
    }}
  />
)

export const Building = () => (
  <BuildStep
    status="in-progress"
    metadata={{
      builds: [
        { component_id: 'c1', component_name: 'api', status: 'success', duration: 2.4 },
        { component_id: 'c2', component_name: 'web', status: 'in-progress' },
      ],
    }}
  />
)

export const Empty = () => <BuildStep status="in-progress" metadata={{}} />
