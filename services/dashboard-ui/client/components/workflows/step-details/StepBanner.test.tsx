import { expect, test } from 'bun:test'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { OrgContext } from '@/providers/org-provider'
import type { TWorkflowStep } from '@/types'
import { StepBanner } from './StepBanner'

const mockOrg = { id: 'org-1', name: 'Acme Corp' } as any

const renderBanner = (step: TWorkflowStep) => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <OrgContext.Provider value={{ org: mockOrg, refresh: () => {} }}>
        <StepBanner step={step} />
      </OrgContext.Provider>
    </QueryClientProvider>
  )
}

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
  renderBanner(
    erroredStep({
      status: 'error',
      status_human_description: 'unable to execute job: exit status 1',
      composite_error: compositeError,
    })
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
  renderBanner(
    erroredStep({
      status: 'error',
      status_human_description: 'unable to execute job: exit status 1',
    })
  )

  expect(screen.getByText('unable to execute job: exit status 1')).toBeDefined()
})

test('surfaces the original error when a failed step was abandoned', () => {
  renderBanner(
    erroredStep({
      status: 'error',
      status_human_description:
        'step abandoned after failure: unable to render terraform variables',
      metadata: {
        abandoned: true,
        original_error:
          'unable to create deploy plan: unable to render terraform variables',
      },
    })
  )

  expect(
    screen.getByText('Step deploy component abandoned after failing')
  ).toBeDefined()
  expect(
    screen.getByText(
      /unable to create deploy plan: unable to render terraform variables/
    )
  ).toBeDefined()
})
