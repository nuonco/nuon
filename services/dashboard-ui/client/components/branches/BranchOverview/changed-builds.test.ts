import { describe, expect, test } from 'bun:test'
import {
  changedBuildRows,
  mergedChangedBuildRows,
  splitChangedBuilds,
} from './changed-builds'

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
        changeReason: undefined,
        componentId: 'cmp_api',
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

describe('mergedChangedBuildRows', () => {
  test('lists changed components from the diff before any build exists', () => {
    const rows = mergedChangedBuildRows({
      orgId: 'org_acme',
      appId: 'app_acme',
      metaBuilds: [],
      runBuilds: [],
      components: [
        {
          name: 'acme-api',
          changeKinds: ['source', 'config'],
          fields: [{ key: 'image' }],
        },
        {
          name: 'acme-web',
          fields: [{ key: 'source' }],
        },
        {
          name: 'acme-worker',
          changeKinds: ['config'],
          fields: [{ key: 'replicas' }],
        },
        {
          name: 'acme-api',
          changeKinds: ['source'],
        },
      ],
    })

    expect(rows).toEqual([
      {
        id: 'acme-api',
        name: 'acme-api',
        status: 'pending',
        changeReason: 'source_and_config',
        kind: 'component',
        componentId: undefined,
        href: undefined,
      },
      {
        id: 'acme-web',
        name: 'acme-web',
        status: 'pending',
        changeReason: 'source_changed',
        kind: 'component',
        componentId: undefined,
        href: undefined,
      },
      {
        id: 'acme-worker',
        name: 'acme-worker',
        status: 'pending',
        changeReason: 'config_changed',
        kind: 'component',
        componentId: undefined,
        href: undefined,
      },
    ])
  })

  test('prefers build status and change reason and keeps uncovered rows', () => {
    const rows = mergedChangedBuildRows({
      orgId: 'org_acme',
      appId: 'app_acme',
      sandboxBuildId: 'sb_acme',
      components: [
        {
          name: 'acme-api',
          changeKinds: ['config'],
        },
      ],
      metaBuilds: [
        {
          component_id: 'cmp_acme_api',
          component_name: 'acme-api',
          status: 'in-progress',
          change_reason: 'source_changed',
        },
        {
          component_id: 'cmp_acme_job',
          component_name: 'acme-job',
          status: 'success',
          change_reason: 'config_changed',
        },
        {
          component_id: 'sandbox',
          component_type: 'sandbox',
          status: 'in-progress',
          change_reason: 'source_changed',
        },
      ],
      runBuilds: [
        {
          id: 'bld_acme_api',
          component_id: 'cmp_acme_api',
          component_name: 'acme-api',
          status: 'queued',
          change_reason: 'config_changed',
        },
      ],
    })

    expect(rows).toEqual([
      {
        id: 'bld_acme_api',
        name: 'acme-api',
        status: 'in-progress',
        changeReason: 'source_changed',
        kind: 'component',
        componentId: 'cmp_acme_api',
        href: '/org_acme/apps/app_acme/components/cmp_acme_api/builds/bld_acme_api',
      },
      {
        id: 'cmp_acme_job',
        name: 'acme-job',
        status: 'success',
        changeReason: 'config_changed',
        kind: 'component',
        componentId: 'cmp_acme_job',
        href: undefined,
      },
      {
        id: 'sb_acme',
        name: 'Sandbox',
        status: 'in-progress',
        changeReason: 'source_changed',
        kind: 'sandbox',
        href: '/org_acme/apps/app_acme/sandbox/builds/sb_acme',
      },
    ])
  })

  test('uses the run build when step metadata is not written yet', () => {
    const rows = mergedChangedBuildRows({
      orgId: 'org_acme',
      appId: 'app_acme',
      metaBuilds: [],
      components: [{ name: 'acme-api', changeKinds: ['source'] }],
      runBuilds: [
        {
          id: 'bld_acme_api',
          component_id: 'cmp_acme_api',
          component_name: 'acme-api',
          status: 'in-progress',
        },
        {
          id: 'bld_acme_worker',
          component_id: 'cmp_acme_worker',
          component_name: 'acme-worker',
          status: 'queued',
          change_reason: 'config_changed',
        },
      ],
    })

    expect(
      rows.map((row) => [row.name, row.status, row.changeReason, row.href])
    ).toEqual([
      [
        'acme-api',
        'in-progress',
        'source_changed',
        '/org_acme/apps/app_acme/components/cmp_acme_api/builds/bld_acme_api',
      ],
      [
        'acme-worker',
        'queued',
        'config_changed',
        '/org_acme/apps/app_acme/components/cmp_acme_worker/builds/bld_acme_worker',
      ],
    ])
  })

  test('leaves build-only rows unchanged when the diff has no components', () => {
    const args = {
      orgId: 'org_acme',
      appId: 'app_acme',
      metaBuilds: [],
      runBuilds: [
        {
          id: 'bld_acme_api',
          component_id: 'cmp_acme_api',
          component_name: 'acme-api',
          status: 'in-progress',
        },
      ],
    }
    expect(mergedChangedBuildRows(args)).toEqual(changedBuildRows(args))
  })
})
