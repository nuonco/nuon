import { DevOrgSection } from './DevOrgSection'

interface IDevOrgSectionContainer {
  orgId: string
}

export const DevOrgSectionContainer = ({
  orgId,
}: IDevOrgSectionContainer) => {
  return <DevOrgSection orgId={orgId} />
}
