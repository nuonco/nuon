import { expect, test } from 'bun:test'
import type { TWorkflow } from '@/types'
import {
  inputChangeSummaries,
  parseWorkflowInputChanges,
} from './WorkflowInputChanges'

test('reads input changes from workflow metadata regardless of workflow type', () => {
  const workflow = {
    type: 'app_branch_config_update',
    metadata: {
      changed_input_values: JSON.stringify({
        api_replicas: { old: '2', new: '4' },
        region: { old: '', new: 'us-west-2' },
      }),
    },
  } as TWorkflow

  expect(parseWorkflowInputChanges(workflow)).toEqual([
    { name: 'api_replicas', old: '2', new: '4' },
    { name: 'region', old: '(empty)', new: 'us-west-2' },
  ])
})

test('turns input changes into summary rows', () => {
  const workflow = {
    metadata: {
      changed_input_values: JSON.stringify({
        api_replicas: { old: '2', new: '4' },
        region: { old: '', new: 'us-west-2' },
      }),
    },
  } as TWorkflow

  expect(inputChangeSummaries(workflow)).toEqual([
    {
      stepId: 'install-inputs',
      stepName: 'Install inputs',
      planType: 'install_input',
      status: 'applied',
      counts: { create: 1, update: 1, delete: 0, replace: 0, noop: 0 },
      countsState: 'ok',
      hasDetail: true,
    },
  ])
})

test('ignores missing or malformed input change metadata', () => {
  expect(
    parseWorkflowInputChanges({
      type: 'input_update',
      metadata: { changed_input_values: '{' },
    } as TWorkflow)
  ).toEqual([])
  expect(
    parseWorkflowInputChanges({ type: 'input_update' } as TWorkflow)
  ).toEqual([])
})
