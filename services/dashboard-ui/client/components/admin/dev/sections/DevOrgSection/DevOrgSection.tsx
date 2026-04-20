import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { DevSection } from '../../shared/DevSection'

interface IDevOrgSection {
  orgId: string
}

export const DevOrgSection = ({ orgId }: IDevOrgSection) => {
  return (
    <DevSection
      title="Development tools"
      subtitle={
        <div className="flex gap-2">
          Org: <ID>{orgId}</ID>
        </div>
      }
    >
      <Text variant="subtext" className="text-gray-500 dark:text-gray-400">
        No dev actions configured yet.
      </Text>
    </DevSection>
  )
}
