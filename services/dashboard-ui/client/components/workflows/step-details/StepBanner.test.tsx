import { expect, test } from 'bun:test'
import { render, screen } from '@testing-library/react'
import type { TWorkflowStep } from '@/types'
import { StepBanner } from './StepBanner'

const compositeError = {
  version: 1,
  type: 'terraform.error',
  severity: 'error',
  message: 'creating S3 Bucket (acme-artifacts): AccessDenied',
  sections: [
    {
      heading: 'Output',
      kind: 'code',
      body: 'not authorized to perform: s3:CreateBucket',
    },
  ],
}

const erroredStep = (status: Record<string, unknown>) =>
  ({
    id: 'step-1',
    name: 'deploy component',
    execution_type: 'system',
    status: { history: [], ...status },
  }) as TWorkflowStep

test('renders the composite error when the step status carries one', () => {
  render(
    <StepBanner
      step={erroredStep({
        status: 'error',
        status_human_description: 'unable to execute job: exit status 1',
        composite_error: compositeError,
      })}
    />
  )

  expect(
    screen.getByText('creating S3 Bucket (acme-artifacts): AccessDenied')
  ).toBeDefined()
  expect(
    screen.getByText('not authorized to perform: s3:CreateBucket')
  ).toBeDefined()
  expect(screen.queryByText('unable to execute job: exit status 1')).toBeNull()
})

test('falls back to the human description when there is no composite error', () => {
  render(
    <StepBanner
      step={erroredStep({
        status: 'error',
        status_human_description: 'unable to execute job: exit status 1',
      })}
    />
  )

  expect(screen.getByText('unable to execute job: exit status 1')).toBeDefined()
})
