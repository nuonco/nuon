export default {
  title: 'Runners/ControlPlaneRecentActivity',
}

import { ControlPlaneRecentActivity } from './ControlPlaneRecentActivity'

const mockActivity = [
  {
    app_id: 'app-1',
    component_id: 'component-1',
    component_name: 'API',
    build_runner_job_id: 'job-1',
    build: {
      id: 'build-1',
      created_at: '2024-01-15T10:00:00Z',
      status: 'active',
    },
  },
  {
    app_id: 'app-1',
    component_id: 'component-2',
    component_name: 'Worker',
    build_runner_job_id: null,
    build: {
      id: 'build-2',
      created_at: '2024-01-14T09:00:00Z',
      status: 'error',
      status_description: 'branch not found',
    },
  },
] as any[]

export const Default = () => (
  <ControlPlaneRecentActivity
    activity={mockActivity}
    orgId="org-1"
    isLoading={false}
    nextCursor="older"
    previousCursor={null}
    onCursorChange={() => {}}
  />
)

export const Loading = () => (
  <ControlPlaneRecentActivity
    activity={[]}
    isLoading={true}
    nextCursor={null}
    previousCursor={null}
    onCursorChange={() => {}}
  />
)

export const Empty = () => (
  <ControlPlaneRecentActivity
    activity={[]}
    isLoading={false}
    nextCursor={null}
    previousCursor={null}
    onCursorChange={() => {}}
  />
)
