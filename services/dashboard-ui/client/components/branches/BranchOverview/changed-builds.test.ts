import { describe, expect, test } from 'bun:test'
import { changedBuildRows } from './changed-builds'

describe('changedBuildRows', () => {
  test('keeps source and config changes and links them to the run build', () => {
    const rows = changedBuildRows({
      orgId: 'org',
      appId: 'app',
      sandboxBuildId: 'sb_1',
      metaBuilds: [
        {
          component_id: 'cmp_api',
          component_name: 'api',
          status: 'success',
          change_reason: 'source_changed',
        },
        {
          component_id: 'cmp_web',
          component_name: 'web',
          status: 'skipped',
          change_reason: 'no_changes',
        },
        {
          component_id: 'cmp_worker',
          component_name: 'worker',
          status: 'in-progress',
          change_reason: 'config_changed',
        },
        {
          component_id: 'sandbox',
          component_type: 'sandbox',
          status: 'success',
          change_reason: 'source_changed',
        },
        {
          component_id: 'cmp_job',
          component_name: 'job',
          status: 'success',
        },
      ],
      runBuilds: [
        { id: 'bld_api', component_id: 'cmp_api' },
        { id: 'bld_worker', component_id: 'cmp_worker' },
        { id: 'bld_job', component_id: 'cmp_job' },
      ],
    })

    expect(rows.map((row) => [row.name, row.href])).toEqual([
      ['api', '/org/apps/app/components/cmp_api/builds/bld_api'],
      ['worker', '/org/apps/app/components/cmp_worker/builds/bld_worker'],
      ['Sandbox', '/org/apps/app/sandbox/builds/sb_1'],
      ['job', '/org/apps/app/components/cmp_job/builds/bld_job'],
    ])
  })

  test('lists run builds when the step has no metadata yet', () => {
    const rows = changedBuildRows({
      orgId: 'org',
      appId: 'app',
      metaBuilds: [],
      runBuilds: [
        {
          id: 'bld_api',
          component_id: 'cmp_api',
          component_name: 'api',
          status: 'in-progress',
        },
      ],
    })
    expect(rows).toEqual([
      {
        id: 'bld_api',
        name: 'api',
        status: 'in-progress',
        href: '/org/apps/app/components/cmp_api/builds/bld_api',
      },
    ])
  })
})
