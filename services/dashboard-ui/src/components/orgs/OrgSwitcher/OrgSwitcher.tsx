'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { FaGithub } from 'react-icons/fa'
import { Plus, TestTube, XCircle } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { ClickToCopy } from '@/components/ClickToCopy'
import { Dropdown } from '@/components/Dropdown'
import { useOrg } from '@/components/Orgs'
import { OrgStatus } from '@/components/OrgStatus'
import { Text } from '@/components/Typography'
import { ConnectGithubModal } from '@/components/vcs-conntections/ConnectGithubModal'
import { VCSConnections } from '@/components/vcs-conntections/VCSConnections'
import type { TOrg } from '@/types'
import { OrgAvatar } from './OrgAvatar'
import { OrgsNav } from './OrgsNav'

export interface IOrgSummary {
  org: TOrg
  shouldPoll?: boolean
  isSidebarOpen?: boolean
}

export const OrgSummary: FC<IOrgSummary> = ({
  org,
  shouldPoll = false,
  isSidebarOpen = true,
}) => {
  return (
    <div className="flex gap-4 items-center justify-start org-summary w-full">
      <OrgAvatar
        name={org?.name}
        logoURL={org?.logo_url}
        isSmall={!isSidebarOpen}
      />

      {isSidebarOpen ? (
        <div className="org-summary-name">
          <Text
            className={classNames(
              'text-md !font-medium leading-normal max-w-[150px] mb-1 break-all text-left !flex-nowrap'
            )}
            title={org?.sandbox_mode ? 'Org is in sandbox mode' : undefined}
          >
            {org?.sandbox_mode && <TestTube className="text-md" />}
            <span
              className={classNames('inline-block truncate', {
                'max-w-[120px]': org?.sandbox_mode,
                'truncate !inline': org?.name?.length >= 16,
              })}
            >
              {org?.name}
            </span>
          </Text>
          <OrgStatus initOrg={org} shouldPoll={shouldPoll} />
        </div>
      ) : null}
    </div>
  )
}

export const OrgVCSConnectionsDetails: FC<{ org: TOrg }> = ({ org }) => {
  return (
    <div className="flex flex-col gap-4 mx-4 py-4 border-cool-grey-600 dark:border-cool-grey-500 border-b border-dotted ">
      <div className="flex items-center justify-between">
        <Text variant="med-14">GitHub connections</Text>
        <ConnectGithubModal />
      </div>

      <div>
        <VCSConnections vcsConnections={org.vcs_connections} />
      </div>
    </div>
  )
}

export interface IOrgSwitcher {
  initOrgs: Array<TOrg>
  isSidebarOpen?: boolean
}

export const OrgSwitcher: FC<IOrgSwitcher> = ({
  initOrgs,
  isSidebarOpen = true,
}) => {
  const { org } = useOrg()
  return (
    <Dropdown
      className={classNames('w-full', {
        '!p-1': !isSidebarOpen,
      })}
      hasCustomPadding
      id="test"
      isFullWidth
      noIcon={!isSidebarOpen}
      text={<OrgSummary org={org} isSidebarOpen={isSidebarOpen} />}
      position="overlay"
      alignment="overlay"
      wrapperClassName="!z-50"
      dropdownContentClassName="min-w-[250px]"
    >
      <div className="flex flex-col gap-4 overflow-auto max-h-[500px] pb-2 overflow-x-hidden">
        <div className="pt-2 px-4 org-details">
          <OrgSummary org={org} />

          <Text className="mt-4" variant="mono-12">
            <ClickToCopy>{org.id}</ClickToCopy>
          </Text>
        </div>
        <OrgVCSConnectionsDetails org={org} />
        <OrgsNav orgs={initOrgs} />
      </div>
    </Dropdown>
  )
}
