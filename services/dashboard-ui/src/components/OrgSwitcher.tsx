'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { FaGithub } from 'react-icons/fa'
import { Plus, TestTube } from '@phosphor-icons/react'
import Image from 'next/image'
import NextLink from 'next/link'
import { setOrgSessionCookie } from '@/app/actions'
import { ClickToCopy } from '@/components/ClickToCopy'
import { Dropdown } from '@/components/Dropdown'
import { Link } from '@/components/Link'
import { useOrg, ConnectGithubModal } from '@/components/Orgs'
import { OrgStatus } from '@/components/OrgStatus'
import { StatusBadge } from '@/components/Status'
import { Text } from '@/components/Typography'
import type { TOrg } from '@/types'
import { GITHUB_APP_NAME, POLL_DURATION, initialsFromString } from '@/utils'
import { SearchInput } from '@/components/SearchInput'

export const OrgAvatar: FC<{
  name: string
  isSmall?: boolean
  logoURL?: string
}> = ({ name, isSmall = false, logoURL }) => {
  return (
    <span
      className={classNames(
        'flex items-center justify-center rounded-md bg-cool-grey-200 text-cool-grey-600 dark:bg-dark-grey-300 dark:text-white/50 font-medium font-sans',
        {
          'w-[40px] h-[40px]': !isSmall,
          'w-[30px] h-[30px]': isSmall,
          'p-2': !logoURL,
        }
      )}
    >
      {logoURL ? (
        <Image
          className="rounded-md"
          height={isSmall ? 30 : 40}
          width={isSmall ? 30 : 40}
          src={logoURL}
          alt="Logo"
        />
      ) : (
        initialsFromString(name)
      )}
    </span>
  )
}

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

const OrgVCSConnections: FC<Pick<TOrg, 'vcs_connections'>> = ({
  vcs_connections,
}) => {
  return (
    <>
      {vcs_connections?.length &&
        vcs_connections?.map((vcs) => (
          <Text
            key={vcs?.id}
            className="flex gap-2 py-4 items-center font-mono text-sm text-cool-grey-600 dark:text-cool-grey-500"
          >
            <FaGithub className="text-lg" /> {vcs?.github_install_id}
          </Text>
        ))}
    </>
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
        <OrgVCSConnections vcs_connections={org.vcs_connections} />
      </div>
    </div>
  )
}

export interface IOrgsNav {
  orgs: Array<TOrg>
}

export const OrgsNav: FC<IOrgsNav> = ({ orgs }) => {
  const [searchTerm, setSearchTerm] = useState<string>('')

  return (
    <div className="flex flex-col gap-4">
      <Text className="px-4" variant="med-14">
        Organizations
      </Text>

      {orgs?.length > 8 ? (
        <div className="px-4 w-full">
          <SearchInput
            className="md:!min-w-full"
            placeholder="Search org name..."
            value={searchTerm}
            onChange={setSearchTerm}
          />
        </div>
      ) : null}

      <nav className="flex flex-col gap-0 px-1">
        {orgs
          .filter((o) =>
            o.name.toLocaleLowerCase().includes(searchTerm.toLocaleLowerCase())
          )
          .map((org) => (
            <NextLink
              className="flex items-center justify-start gap-4 rounded-md p-2 hover:bg-cool-grey-600/20"
              key={org.id}
              href={`/${org.id}/apps`}
            >
              <OrgAvatar name={org.name} logoURL={org.logo_url} />
              <span>
                <Text
                  className="break-all text-md font-medium leading-normal mb-1 !flex-nowrap"
                  title={
                    org.sandbox_mode ? 'Org is in sandbox mode' : undefined
                  }
                >
                  {org.sandbox_mode && <TestTube className="text-sm" />}
                  <span
                    className={classNames('', {
                      'truncate !inline max-w-[140px]': org.name.length >= 16,
                    })}
                  >
                    {org.name}
                  </span>
                </Text>
                <StatusBadge status={org.status} isWithoutBorder />
              </span>
            </NextLink>
          ))}
      </nav>
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
  //  const [org, updateOrg] = useState<TOrg>(initOrg)

  /* useEffect(() => {
   *   async function setSession() {
   *     await setOrgSessionCookie(initOrg.id)
   *   }

   *   setSession()

   *   const fetchOrg = () => {
   *     fetch(`/api/${initOrg.id}`)
   *       .then((res) =>
   *         res.json().then((o) => {
   *           updateOrg(o)
   *         })
   *       )
   *       .catch(console.error)
   *   }

   *   const pollOrg = setInterval(fetchOrg, POLL_DURATION)
   *   return () => clearInterval(pollOrg)
   * }, []) */

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
