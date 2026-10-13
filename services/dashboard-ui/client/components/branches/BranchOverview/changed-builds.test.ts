import { describe, expect, test } from 'bun:test'
import { changedBuildRows, splitChangedBuilds } from './changed-builds'

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

    expect(rows.map((row) => [row.name, row.changeReason, row.href])).toEqual([
      [
        'api',
        'source_changed',
        '/org/apps/app/components/cmp_api/builds/bld_api',
      ],
      [
        'worker',
        'config_changed',
        '/org/apps/app/components/cmp_worker/builds/bld_worker',
      ],
      ['Sandbox', 'source_changed', '/org/apps/app/sandbox/builds/sb_1'],
      [
        'job',
        'source_changed',
        '/org/apps/app/components/cmp_job/builds/bld_job',
      ],
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
        kind: 'component',
        href: '/org/apps/app/components/cmp_api/builds/bld_api',
      },
    ])
  })

  test('puts a build that changed both source and config in both cards', () => {
    const rows = changedBuildRows({
      orgId: 'org',
      appId: 'app',
      metaBuilds: [
        {
          component_id: 'cmp_alb',
          component_name: 'alb',
          status: 'success',
          change_reason: 'source_changed',
        },
        {
          component_id: 'cmp_nginx',
          component_name: 'img_nginx',
          status: 'success',
          change_reason: 'source_and_config',
        },
        {
          component_id: 'cmp_web',
          component_name: 'web',
          status: 'skipped',
          change_reason: 'no_changes',
        },
      ],
      runBuilds: [
        { id: 'bld_alb', component_id: 'cmp_alb' },
        { id: 'bld_nginx', component_id: 'cmp_nginx' },
      ],
    })

    const cards = splitChangedBuilds(rows)
    expect(cards.config.map((row) => row.name)).toEqual(['img_nginx'])
    expect(cards.source.map((row) => row.name)).toEqual(['alb', 'img_nginx'])
  })
})
