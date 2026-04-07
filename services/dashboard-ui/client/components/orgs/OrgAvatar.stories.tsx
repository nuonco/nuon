export default {
  title: 'Orgs/OrgAvatar',
}

import { OrgAvatar } from './OrgAvatar'

export const Small = () => <OrgAvatar name="Acme Corp" size="sm" />
export const Medium = () => <OrgAvatar name="Acme Corp" size="md" />
export const Large = () => <OrgAvatar name="Acme Corp" size="lg" />
export const ExtraLarge = () => <OrgAvatar name="Acme Corp" size="xl" />

export const LongName = () => <OrgAvatar name="Some Very Long Organization Name" size="xl" />

export const NoName = () => <OrgAvatar size="md" />
