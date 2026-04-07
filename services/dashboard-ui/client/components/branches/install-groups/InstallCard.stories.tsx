import { InstallCard } from './InstallCard'

export default { title: 'Branches/InstallGroups/InstallCard' }

export const Default = () => (
  <InstallCard
    install={{ id: 'inst-1', name: 'Production US East' } as any}
  />
)

export const Disabled = () => (
  <InstallCard
    install={{ id: 'inst-2', name: 'Staging' } as any}
    isDisabled
  />
)
