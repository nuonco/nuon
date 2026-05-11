export default {
  title: 'Onboarding/V2 Steps/WelcomeNameOrgStep',
}

import type { TOrg } from '@/types'
import { WelcomeNameOrgStep } from './WelcomeNameOrgStep'

const mockOrgs = [
  { id: 'org01jdr5xkxxgenrk2pxn1kdv', name: 'swift-harbor-ridge' },
  { id: 'org02k4q9wzzbpemflp7sn1abc', name: 'acme-corp' },
  { id: 'org03zh3qvyyhxqemxk4uo2def', name: 'sandbox-team' },
] as TOrg[]

export const Default = () => (
  <WelcomeNameOrgStep
    orgName=""
    setOrgName={() => {}}
    isPending={false}
    waiting={false}
    displayName=""
    onSubmit={(e) => e.preventDefault()}
    onAdvance={() => {}}
    onGenerateName={() => {}}
  />
)

export const LoadingExistingOrgs = () => (
  <WelcomeNameOrgStep
    orgName=""
    setOrgName={() => {}}
    isPending={false}
    waiting={false}
    displayName=""
    isExistingOrgsLoading
    onSelectExistingOrg={() => {}}
    onSubmit={(e) => e.preventDefault()}
    onAdvance={() => {}}
    onGenerateName={() => {}}
  />
)

export const WithExistingOrgs = () => (
  <WelcomeNameOrgStep
    orgName=""
    setOrgName={() => {}}
    isPending={false}
    waiting={false}
    displayName=""
    existingOrgs={mockOrgs}
    onSelectExistingOrg={() => {}}
    onSubmit={(e) => e.preventDefault()}
    onAdvance={() => {}}
    onGenerateName={() => {}}
  />
)

export const AttachingExistingOrg = () => (
  <WelcomeNameOrgStep
    orgName=""
    setOrgName={() => {}}
    isPending={false}
    waiting={false}
    displayName=""
    existingOrgs={mockOrgs}
    attachingOrgId={mockOrgs[1].id}
    isAttachPending
    onSelectExistingOrg={() => {}}
    onSubmit={(e) => e.preventDefault()}
    onAdvance={() => {}}
    onGenerateName={() => {}}
  />
)

export const WithAttachedOrg = () => (
  <WelcomeNameOrgStep
    org={mockOrgs[0]}
    orgName=""
    setOrgName={() => {}}
    isPending={false}
    waiting={false}
    displayName={mockOrgs[0].name!}
    existingOrgs={[mockOrgs[0]]}
    onSelectExistingOrg={() => {}}
    onSubmit={(e) => e.preventDefault()}
    onAdvance={() => {}}
    onGenerateName={() => {}}
  />
)

export const WithAttachedOrgAndAlternatives = () => (
  <WelcomeNameOrgStep
    org={mockOrgs[1]}
    orgName=""
    setOrgName={() => {}}
    isPending={false}
    waiting={false}
    displayName={mockOrgs[1].name!}
    existingOrgs={mockOrgs}
    onSelectExistingOrg={() => {}}
    onSubmit={(e) => e.preventDefault()}
    onAdvance={() => {}}
    onGenerateName={() => {}}
  />
)
