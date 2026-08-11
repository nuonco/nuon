export default {
  title: 'Onboarding/V1 Steps/CreateOrgStep',
}

import { CreateOrgStep, CompletedOrgCard } from './CreateOrgStep'
import type { TOrg } from '@/types'

const mockOrg = {
  id: 'org-123',
  name: 'swift-harbor-ridge',
  status: 'active',
} as TOrg

const generateName = () => Promise.resolve('random-generated-name')

export const Default = () => (
  <CreateOrgStep
    onAdvance={() => alert('Advancing to next step')}
    nextStepTitle="Set up your app"
    createdOrg={null}
    isPending={false}
    error={null}
    onCreateOrg={(name) => alert(`Creating org: ${name}`)}
    onGenerateName={generateName}
  />
)

export const Pending = () => (
  <CreateOrgStep
    onAdvance={() => {}}
    nextStepTitle="Set up your app"
    createdOrg={null}
    isPending={true}
    error={null}
    onCreateOrg={() => {}}
    onGenerateName={generateName}
  />
)

export const WithError = () => (
  <CreateOrgStep
    onAdvance={() => {}}
    nextStepTitle="Set up your app"
    createdOrg={null}
    isPending={false}
    error={{
      error: 'Organization name already taken.',
      description: '',
      user_error: true,
    }}
    onCreateOrg={() => {}}
    onGenerateName={generateName}
  />
)

export const Created = () => (
  <CreateOrgStep
    onAdvance={() => alert('Advancing to next step')}
    nextStepTitle="Set up your app"
    createdOrg={mockOrg}
    isPending={false}
    error={null}
    onCreateOrg={() => {}}
    onGenerateName={generateName}
  />
)

export const CompletedOrgLoading = () => (
  <CompletedOrgCard
    org={undefined}
    orgId="org-123"
    isLoading={true}
    onAdvance={() => {}}
    nextStepTitle="Set up your app"
  />
)

export const CompletedOrgActive = () => (
  <CompletedOrgCard
    org={mockOrg}
    orgId="org-123"
    isLoading={false}
    onAdvance={() => alert('Advancing to next step')}
    nextStepTitle="Set up your app"
  />
)

export const CompletedOrgProvisioning = () => (
  <CompletedOrgCard
    org={{ ...mockOrg, status: 'provisioning' } as TOrg}
    orgId="org-123"
    isLoading={false}
    onAdvance={() => alert('Advancing to next step')}
    nextStepTitle="Set up your app"
  />
)
