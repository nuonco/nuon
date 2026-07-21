import { describe, expect, test } from 'bun:test'
import {
  compileBlock,
  compileTemplate,
  getArraySources,
  getStateVariables,
  previewBlock,
  previewDocument,
  resolvePath,
  substituteVariables,
} from './compiler'
import type { TBannerBlock, TBlock, TTableBlock } from './types'

const state = {
  app: { name: 'control-plane' },
  install: { name: 'prod-cluster', status: 'active' },
  sandbox: { status: 'finished', outputs: { cluster: { version: '1.31' } } },
  secrets: { token: 'hidden' },
  cloud_account: { id: 'aws-123' },
  actions: {
    workflows: {
      recent: {
        outputs: {
          workflows: [
            { name: 'deploy', status: 'finished', started_at: '2026-07-20T10:00:00Z' },
            { name: 'sync', status: 'error', started_at: '2026-07-19T10:00:00Z' },
            { name: 'plan', status: 'finished', started_at: '2026-07-18T10:00:00Z' },
          ],
        },
      },
    },
  },
}

const tableBlock: TTableBlock = {
  key: 't1',
  type: 'table',
  sourcePath: 'actions.workflows.recent.outputs.workflows',
  limit: 2,
  emptyText: 'No workflows yet',
  columns: [
    { key: 'c1', header: 'Status', kind: 'status', path: 'status' },
    { key: 'c2', header: 'Name', kind: 'text', path: 'name' },
    { key: 'c3', header: 'Started', kind: 'time', path: 'started_at' },
  ],
}

const bannerBlock: TBannerBlock = {
  key: 'b1',
  type: 'banner',
  theme: 'warn',
  content: 'Setup still required for {{.nuon.install.name}}.',
  condition: { path: 'sandbox.status', op: 'ne', value: 'finished' },
}

describe('compileBlock', () => {
  test('markdown passes through', () => {
    const block: TBlock = { key: 'm', type: 'markdown', content: '# Hi\n' }
    expect(compileBlock(block, 0)).toBe('# Hi')
  })

  test('banner wraps content in condition', () => {
    const output = compileBlock(bannerBlock, 0)
    expect(output).toContain(
      '{{ if ne (printf "%v" (dig "sandbox" "status" "" .nuon)) "finished" }}'
    )
    expect(output).toContain('<nuon-banner theme="warn">')
    expect(output).toContain('{{ end }}')
  })

  test('banner without condition has no if', () => {
    const output = compileBlock({ ...bannerBlock, condition: undefined }, 0)
    expect(output).not.toContain('{{ if')
  })

  test('runbook block compiles to nuon-run-runbook tag', () => {
    const block: TBlock = { key: 'r', type: 'runbook', id: 'rbk-1', name: 'Rotate credentials' }
    expect(compileBlock(block, 0)).toBe(
      '<nuon-run-runbook id="rbk-1" name="Rotate credentials"></nuon-run-runbook>'
    )
  })

  test('action block compiles to nuon-action-card tag', () => {
    const block: TBlock = { key: 'a', type: 'action', id: 'act-1', name: 'restart-api' }
    expect(compileBlock(block, 0)).toBe(
      '<nuon-action-card id="act-1" name="restart-api"></nuon-action-card>'
    )
  })

  test('component block compiles to nuon-component-card tag', () => {
    const block: TBlock = { key: 'c', type: 'component', id: 'cmp-1', name: 'ctl-api' }
    expect(compileBlock(block, 0)).toBe(
      '<nuon-component-card id="cmp-1" name="ctl-api"></nuon-component-card>'
    )
  })

  test('entity block escapes attribute values', () => {
    const block: TBlock = { key: 'r', type: 'runbook', id: '', name: 'a "quoted" <name>' }
    expect(compileBlock(block, 0)).toBe(
      '<nuon-run-runbook name="a &#34;quoted&#34; &lt;name&gt;"></nuon-run-runbook>'
    )
  })

  test('entity block without selection compiles to nothing', () => {
    const block: TBlock = { key: 'r', type: 'runbook', id: '', name: '' }
    expect(compileBlock(block, 0)).toBe('')
  })

  test('table emits dig source, slice limit, empty state, and typed cells', () => {
    const output = compileBlock(tableBlock, 3)
    expect(output).toContain(
      '(dig "actions" "workflows" "recent" "outputs" "workflows" (list) .nuon)'
    )
    expect(output).toContain('{{ if gt (len $rows3) 2 }}{{ $rows3 = slice $rows3 0 2 }}{{ end }}')
    expect(output).toContain('_No workflows yet_')
    expect(output).toContain('{{ range $row3 := $rows3 }}')
    expect(output).toContain('<nuon-status status="{{ (dig "status" "" $row3) }}"')
    expect(output).toContain('<nuon-time time="{{ (dig "started_at" "" $row3) }}"')
    expect(output).toContain('{{ (dig "name" "—" $row3) }}')
  })

  test('status row renders labelled values', () => {
    const output = compileBlock(
      {
        key: 's',
        type: 'status-row',
        items: [
          { key: 'i1', label: 'Status', kind: 'status', path: 'install.status' },
          { key: 'i2', label: 'Name', kind: 'text', path: 'install.name' },
        ],
      },
      0
    )
    expect(output).toContain('<nuon-status status="{{ (dig "install" "status" "" .nuon) }}"')
    expect(output).toContain('{{ (dig "install" "name" "—" .nuon) }}')
  })
})

describe('compileTemplate', () => {
  test('joins blocks and drops empty ones', () => {
    const output = compileTemplate([
      { key: 'a', type: 'markdown', content: '# Title' },
      { key: 'b', type: 'markdown', content: '   ' },
      { key: 'c', type: 'raw', content: '{{ $x := 1 }}' },
    ])
    expect(output).toStartWith('{{/* Generated with the Nuon README studio */}}')
    expect(output).toContain('# Title\n\n{{ $x := 1 }}')
    expect(output).toEndWith('\n')
  })
})

describe('preview', () => {
  test('substitutes scalar variables and escapes html', () => {
    expect(
      substituteVariables('Version {{.nuon.sandbox.outputs.cluster.version}}', state)
    ).toBe('Version 1.31')
    expect(substituteVariables('{{.nuon.missing.path}}', state)).toBe(
      '{{.nuon.missing.path}}'
    )
  })

  test('banner hidden when condition is false', () => {
    expect(previewBlock(bannerBlock, state)).toBe('')
    expect(
      previewBlock(
        { ...bannerBlock, condition: { path: 'sandbox.status', op: 'eq', value: 'finished' } },
        state
      )
    ).toContain('Setup still required for prod-cluster.')
  })

  test('table renders limited rows from state', () => {
    const output = previewBlock(tableBlock, state)
    expect(output).toContain('<nuon-status status="finished"')
    expect(output).toContain('<td>deploy</td>')
    expect(output).toContain('<td>sync</td>')
    expect(output).not.toContain('<td>plan</td>')
  })

  test('table falls back to empty text without state', () => {
    expect(previewBlock(tableBlock, undefined)).toBe('_No workflows yet_')
  })

  test('entity block previews as a card stand-in', () => {
    const output = previewBlock(
      { key: 'r', type: 'runbook', id: 'rbk-1', name: 'Rotate credentials' },
      state
    )
    expect(output).toContain('Run runbook: Rotate credentials')
    expect(output).toContain('install page')
  })

  test('entity block without selection previews a placeholder', () => {
    const output = previewBlock(
      { key: 'c', type: 'component', id: '', name: '' },
      state
    )
    expect(output).toContain('select a component')
  })

  test('previewDocument joins non-empty sections', () => {
    const output = previewDocument(
      [{ key: 'a', type: 'markdown', content: '# T' }, bannerBlock],
      state
    )
    expect(output).toBe('# T')
  })
})

describe('state helpers', () => {
  test('resolvePath walks nested objects', () => {
    expect(resolvePath(state, 'sandbox.outputs.cluster.version')).toBe('1.31')
    expect(resolvePath(state, 'nope.nope')).toBeUndefined()
  })

  test('getStateVariables excludes secrets and cloud_account', () => {
    const variables = getStateVariables(state)
    const templates = variables.map((variable) => variable.template)
    expect(templates).toContain('{{.nuon.install.name}}')
    expect(templates.some((template) => template.includes('secrets'))).toBe(false)
    expect(templates.some((template) => template.includes('cloud_account'))).toBe(false)
  })

  test('getArraySources finds arrays of objects with scalar keys', () => {
    const sources = getArraySources(state)
    expect(sources).toEqual([
      {
        path: 'actions.workflows.recent.outputs.workflows',
        keys: ['name', 'status', 'started_at'],
        length: 3,
      },
    ])
  })
})
