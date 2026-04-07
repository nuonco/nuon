export default {
  title: 'Onboarding/V2 Steps/WelcomeNameOrgStep',
}

import { WelcomeNameOrgStep } from './WelcomeNameOrgStep'

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

export const WithOrg = () => (
  <WelcomeNameOrgStep
    org={{ id: 'org-1', name: 'swift-harbor-ridge' } as any}
    orgName="swift-harbor-ridge"
    setOrgName={() => {}}
    isPending={false}
    waiting={false}
    displayName="swift-harbor-ridge"
    onSubmit={(e) => e.preventDefault()}
    onAdvance={() => {}}
    onGenerateName={() => {}}
  />
)
