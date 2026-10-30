import { afterEach, expect, test } from 'bun:test'
import { useState } from 'react'
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { ConnectionForm } from './ConnectionForm'
import { AddRepositoryAccess } from '../AddRepositoryAccess/AddRepositoryAccess'
import { ecrConnection } from '../__stories__/fixtures'
import type { TConnectionValues } from './schema'

afterEach(cleanup)

for (const scenario of [
  {
    platform: 'aws',
    values: {
      name: 'acme-ecr',
      target_id: '123456789012',
      role_name: 'acme-images',
      ecr_region: 'eu-west-1',
    },
    expected: {
      name: 'acme-ecr',
      platform: 'aws',
      target_id: '123456789012',
      preset: 'ecr',
      principal: 'arn:aws:iam::123456789012:role/acme-images',
      default_region: 'eu-west-1',
    },
    absent: ['Project ID', 'Subscription ID'],
  },
  {
    platform: 'gcp',
    values: {
      name: 'acme-gar',
      platform: 'gcp',
      target_id: 'acme-project',
      gcp_project_number: '987654321012',
      service_account_name: 'acme-images',
      gar_location: 'europe',
    },
    expected: {
      name: 'acme-gar',
      platform: 'gcp',
      target_id: 'acme-project',
      preset: 'gar',
      principal: 'acme-images@acme-project.iam.gserviceaccount.com',
      gcp_project_number: '987654321012',
      default_region: 'europe',
    },
    absent: [
      'AWS account ID',
      'IAM role name',
      'ECR region',
      'Subscription ID',
    ],
  },
  {
    platform: 'azure',
    values: {
      name: 'acme-acr',
      platform: 'azure',
      target_id: '11111111-1111-4111-8111-111111111111',
      azure_tenant_id: '22222222-2222-4222-8222-222222222222',
      principal: '33333333-3333-4333-8333-333333333333',
      azure_resource_group: 'acme-images',
      azure_registry_name: 'acmeimages',
      azure_registry_login_server: 'acmeimages-abc123.azurecr.io',
    },
    expected: {
      name: 'acme-acr',
      platform: 'azure',
      target_id: '11111111-1111-4111-8111-111111111111',
      preset: 'acr',
      principal: '33333333-3333-4333-8333-333333333333',
      azure_tenant_id: '22222222-2222-4222-8222-222222222222',
      azure_registry_id:
        '/subscriptions/11111111-1111-4111-8111-111111111111/resourceGroups/acme-images/providers/Microsoft.ContainerRegistry/registries/acmeimages',
      azure_registry_login_server: 'acmeimages-abc123.azurecr.io',
    },
    absent: [
      'AWS account ID',
      'IAM role name',
      'ECR region',
      'Project ID',
      'GAR location',
    ],
  },
] as const) {
  test(`${scenario.platform} creates only saved identity and target details, without repository inputs`, async () => {
    let submitted: unknown
    render(
      <MemoryRouter>
        <ConnectionForm
          values={scenario.values as Partial<TConnectionValues>}
          onCreate={(body) => {
            submitted = body
          }}
        />
      </MemoryRouter>
    )
    for (const label of scenario.absent)
      expect(screen.queryByLabelText(label)).toBeNull()
    expect(screen.queryByText('Repository names')).toBeNull()
    expect(screen.queryByRole('button', { name: 'Add repository' })).toBeNull()
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Create connection' }))
    })
    expect(submitted).toEqual(scenario.expected)
  })
}

for (const preset of ['stacks', 'custom'] as const) {
  test(`${preset} connections still create without image registry fields`, async () => {
    let submitted: unknown
    render(
      <MemoryRouter>
        <ConnectionForm
          values={{
            name: 'acme-stack',
            preset,
            target_id: '123456789012',
            role_name: 'acme-stack',
          }}
          onCreate={(body) => {
            submitted = body
          }}
        />
      </MemoryRouter>
    )
    expect(screen.queryByLabelText('ECR region')).toBeNull()
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Create connection' }))
    })
    expect(submitted).toEqual({
      name: 'acme-stack',
      platform: 'aws',
      preset,
      target_id: '123456789012',
      principal: 'arn:aws:iam::123456789012:role/acme-stack',
    })
  })
}

test('switching from AWS to GAR and ACR validates only the selected cloud and omits stale AWS fields', async () => {
  let submitted: unknown
  render(
    <MemoryRouter>
      <ConnectionForm
        values={{
          name: 'acme-images',
          target_id: '123456789012',
          ecr_region: 'us-west-2',
        }}
        onCreate={(body) => {
          submitted = body
        }}
      />
    </MemoryRouter>
  )
  fireEvent.click(screen.getByLabelText('Cloud platform'))
  fireEvent.click(screen.getByRole('option', { name: 'Google Cloud' }))
  expect(screen.queryByLabelText('AWS account ID')).toBeNull()
  expect(screen.getByLabelText('Project ID')).toHaveValue('')
  expect(
    screen.queryByText(
      'Use a 6–30 character GCP project ID with lowercase letters, numbers, and dashes'
    )
  ).toBeNull()
  await act(async () => {
    fireEvent.change(screen.getByLabelText('Project ID'), {
      target: { value: 'acme-project' },
    })
    fireEvent.change(screen.getByLabelText('Project number'), {
      target: { value: '987654321012' },
    })
    fireEvent.change(screen.getByLabelText('GAR location'), {
      target: { value: 'us' },
    })
  })
  await act(async () => {
    fireEvent.click(screen.getByRole('button', { name: 'Create connection' }))
  })
  expect(submitted).toEqual({
    name: 'acme-images',
    platform: 'gcp',
    target_id: 'acme-project',
    preset: 'gar',
    principal: 'nuon-cloud-connection@acme-project.iam.gserviceaccount.com',
    gcp_project_number: '987654321012',
    default_region: 'us',
  })
  fireEvent.click(screen.getByLabelText('Cloud platform'))
  fireEvent.click(screen.getByRole('option', { name: 'Azure' }))
  expect(screen.queryByLabelText('GAR location')).toBeNull()
  expect(screen.queryByLabelText('ECR region')).toBeNull()
  expect(screen.getByLabelText('Registry login server')).toBeVisible()
  expect(
    screen.queryByText(
      'Enter a UUID in the form 00000000-0000-0000-0000-000000000000'
    )
  ).toBeNull()
  expect(
    screen.getByRole('button', { name: 'Create connection' })
  ).toHaveAttribute('aria-disabled', 'true')
})

test('repository generator shows policy and commands for edited rows and hides stale output', async () => {
  const Generator = () => {
    const [result, setResult] = useState<{
      names: string[]
      permissionsCli: string
      permissionsPolicy: Record<string, unknown>
    }>()
    return (
      <AddRepositoryAccess
        connection={ecrConnection}
        onGenerate={(names) =>
          setResult({
            names,
            permissionsCli: `grant ${names.join(' ')}`,
            permissionsPolicy: { repositories: names },
          })
        }
        result={result}
      />
    )
  }
  render(
    <MemoryRouter>
      <Generator />
    </MemoryRouter>
  )
  expect(screen.queryByRole('textbox')).toBeNull()
  fireEvent.click(
    screen.getByRole('button', { name: 'Generate repository access commands' })
  )
  await act(async () => {
    fireEvent.change(screen.getByLabelText('ECR repository name 1'), {
      target: { value: 'acme/api' },
    })
  })
  await act(async () => {
    fireEvent.click(screen.getByRole('button', { name: 'Generate commands' }))
  })
  expect(
    screen.getByRole('region', { name: 'Generated repository access' })
  ).toHaveTextContent('acme/api')
  fireEvent.click(screen.getByRole('button', { name: /^Commands$/ }))
  expect(
    screen.getByRole('region', { name: 'Generated repository access' })
  ).toHaveTextContent('grant acme/api')
  await act(async () => {
    fireEvent.change(screen.getByLabelText('ECR repository name 1'), {
      target: { value: 'acme/worker' },
    })
  })
  expect(
    screen.queryByRole('region', { name: 'Generated repository access' })
  ).toBeNull()
})
