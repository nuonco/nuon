import { describe, expect, test } from 'bun:test'
import {
  importRunbook,
  serializeRunbook,
  serializeRunbookReadme,
  validateRunbook,
} from './helpers'

describe('runbook builder helpers', () => {
  test('serializes supported fields and escapes strings', () => {
    const toml = serializeRunbook('release "one"', '', [
      {
        key: '1',
        operation: 'check-component-drift',
        name: 'Check API drift',
        componentName: 'api',
        deployDependents: true,
      },
    ])
    expect(toml).toContain('name = "release \\"one\\""')
    expect(toml).toContain(
      'type = "component_deploy"\nplan_only = true\ncomponent_name = "api"\ndeploy_dependents = true'
    )
  })
  test('generates a companion readme from step documentation', () => {
    const steps = [
      {
        key: '1',
        operation: 'check-component-drift' as const,
        name: 'Check API drift',
        documentation: 'Review the plan before continuing.\n\n> This step does not apply changes.',
        componentName: 'api',
      },
      {
        key: '2',
        operation: 'deploy-component' as const,
        name: 'Deploy API',
        componentName: 'api',
      },
    ]
    expect(serializeRunbook('Release API', '', steps)).toContain(
      'readme = "./runbooks/release-api.md"'
    )
    expect(serializeRunbookReadme('Release API', 'Release workflow.', steps)).toBe(
      '# Release API\n\nRelease workflow.\n\n## 1. Check API drift\n\n**Operation:** Check component drift\n\nReview the plan before continuing.\n\n> This step does not apply changes.\n\n## 2. Deploy API\n\n**Operation:** Deploy component\n'
    )
  })
  test('does not add a readme reference without step documentation', () => {
    expect(
      serializeRunbook('Release API', '', [
        {
          key: '1',
          operation: 'deploy-component',
          name: 'Deploy API',
          componentName: 'api',
        },
      ])
    ).not.toContain('readme =')
  })
  test('validates required settings', () => {
    expect(validateRunbook('', [{ key: '1', operation: 'command', name: 'Run command' }])).toEqual([
      'Enter a runbook name.',
      'Step 1 requires a command.',
    ])
  })
  test('rejects an invalid command timeout', () => {
    expect(
      validateRunbook('verify', [
        { key: '1', operation: 'command', name: 'Verify', command: 'curl localhost', timeout: 'five minutes' },
      ])
    ).toEqual(['Step 1 requires a valid timeout such as 30s or 5m.'])
  })
  test('rejects a command timeout that overflows a Go duration', () => {
    const command = { key: '1', operation: 'command' as const, name: 'Verify', command: 'true' }
    expect(validateRunbook('verify', [{ ...command, timeout: '2562047h47m16.854775807s' }])).toEqual([])
    expect(validateRunbook('verify', [{ ...command, timeout: '2562047h47m16.854775808s' }])).toEqual([
      'Step 1 requires a valid timeout such as 30s or 5m.',
    ])
    expect(validateRunbook('verify', [{ ...command, timeout: '-2562047h47m16.854775808s' }])).toEqual([])
  })
  test('imports latest steps and maps action ids', () => {
    const result = importRunbook(
      {
        id: 'r',
        name: 'Existing',
        configs: [
          {
            id: 'c',
            steps: [
              { id: 's', name: 'act', type: 'action', action_workflow_id: 'a' },
            ],
          },
        ],
      },
      [{ id: 'a', name: 'backup' }]
    )
    expect(result.errors).toEqual([])
    expect(result.steps[0]?.name).toBe('act')
    expect(result.steps[0]?.actionName).toBe('backup')
  })
  test('skips actions that cannot be mapped', () => {
    const result = importRunbook(
      {
        id: 'r',
        name: 'Existing',
        configs: [
          {
            id: 'c',
            steps: [
              {
                id: 's',
                name: 'act',
                type: 'action',
                action_workflow_id: 'missing',
              },
            ],
          },
        ],
      },
      []
    )
    expect(result.steps).toEqual([])
    expect(result.errors[0]).toContain('could not map')
  })

  test('rejects configured actions mixed with inline action fields', () => {
    const result = importRunbook(
      {
        id: 'r',
        name: 'Existing',
        configs: [
          {
            id: 'c',
            steps: [
              {
                id: 's',
                name: 'act',
                type: 'action',
                action_workflow_id: 'a',
                command: 'echo different behavior',
              },
            ],
          },
        ],
      },
      [{ id: 'a', name: 'backup' }]
    )
    expect(result.steps).toEqual([])
    expect(result.errors[0]).toContain('mixes a configured action')
  })

  test('does not partially import unsupported steps', () => {
    const result = importRunbook(
      {
        id: 'r',
        name: 'Existing',
        configs: [
          {
            id: 'c',
            steps: [
              { id: 'one', name: 'deploy', type: 'component_deploy', component_name: 'api' },
              { id: 'two', name: 'script', type: 'action', inline_contents: './verify.sh' },
            ],
          },
        ],
      },
      []
    )
    expect(result.steps).toEqual([])
    expect(result.errors[0]).toContain('inline script contents')
  })

  test('converts imported timeout nanoseconds to a Go duration', () => {
    const result = importRunbook(
      {
        id: 'r',
        name: 'Existing',
        configs: [
          {
            id: 'c',
            steps: [
              { id: 's', name: 'verify', type: 'action', command: 'curl localhost', timeout: 300_000_000_000 },
            ],
          },
        ],
      },
      []
    )
    expect(result.steps[0]?.timeout).toBe('5m')
  })
})
