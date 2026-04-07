export default {
  title: 'Onboarding/V1 Steps/CreateInstallStep',
}

import { CompletedInstallCard } from './CreateInstallStep'

export const CompletedCard = () => (
  <CompletedInstallCard
    install={{ name: 'Demo Install' } as any}
    installId="install-1"
    orgId="org-1"
    isLoading={false}
  />
)

export const CompletedCardLoading = () => (
  <CompletedInstallCard
    install={undefined}
    installId="install-1"
    orgId="org-1"
    isLoading={true}
  />
)
