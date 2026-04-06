import { Avatar } from '@/components/common/Avatar'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { useSidebar } from '@/hooks/use-sidebar'
import type { TOrg } from '@/types'
import { cn } from '@/utils/classnames'

const OrgAvatar = ({
  isButtonSummary = false,
  org,
}: {
  isButtonSummary?: boolean
  org: TOrg
}) => {
  const { isSidebarOpen } = useSidebar()
  return (
    <div className={cn({ 'org-avatar-summary relative': isButtonSummary })}>
      <Avatar
        {...(org?.logo_url ? { src: org?.logo_url } : { name: org.name })}
        size={!isSidebarOpen && isButtonSummary ? 'md' : 'xl'}
      />
      {isButtonSummary ? (
        <Status
          className={cn('absolute right-0 top-0 transition-all', {
            'opacity-0': isSidebarOpen,
            'delay-fastest opacity-100': !isSidebarOpen,
          })}
          status={org?.status}
          isWithoutText
        />
      ) : null}
    </div>
  )
}

export interface IOrgSummary {
  isSidebarOpen?: boolean
  isButtonSummary?: boolean
  org: TOrg
}

export const OrgSummary = ({
  isSidebarOpen = true,
  isButtonSummary = false,
  org,
}: IOrgSummary) => {
  return (
    <div className="flex gap-4 items-center overflow-hidden">
      <OrgAvatar {...{ isButtonSummary, org }} />
      <div
        className={cn('transition-all max-w-full overflow-hidden', {
          'md:opacity-100': isSidebarOpen,
          'md:opacity-0': !isSidebarOpen,
        })}
      >
        <Text
          weight="strong"
          variant="subtext"
          flex
          className="text-nowrap"
        >
          {org.sandbox_mode && (
            <Icon
              variant="TestTube"
              className="!w-[12px] !h-[12px] shrink-0"
              size="12"
            />
          )}
          <span className="truncate">{org.name}</span>
        </Text>
        <Status status={org?.status} />
      </div>
    </div>
  )
}
