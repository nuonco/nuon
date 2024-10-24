'use client'

import classNames from 'classnames'
import React, { type FC, useEffect } from 'react'
import { FaGithub } from 'react-icons/fa'
import { Plus, TestTube } from '@phosphor-icons/react'
import NextLink from 'next/link'
import { setOrgSessionCookie } from '@/app/actions'
import { ClickToCopy } from '@/components/ClickToCopy'
import { Dropdown } from '@/components/Dropdown'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { Text } from '@/components/Typography'
import type { TOrg } from '@/types'
import { GITHUB_APP_NAME, initialsFromString } from '@/utils'

export const OrgAvatar: FC<{ name: string; isSmall?: boolean }> = ({
  name,
  isSmall = false,
}) => {
  return (
    <span
      className={classNames(
        'flex items-center justify-center p-2 rounded-md bg-cool-grey-200 text-cool-grey-600 dark:bg-dark-grey-300 dark:text-white/50 font-medium font-sans',
        {
          'w-[40px] h-[40px]': !isSmall,
          'w-[30px] h-[30px]': isSmall,
        }
      )}
    >
      {initialsFromString(name)}
    </span>
  )
}

// NOTE(nnnnat): new semiflat designed org switcher parts
export interface IOrgSummary {
  org: TOrg
}

export const OrgSummary: FC<IOrgSummary> = ({ org }) => {
  return (
    <div className="flex gap-4 items-center justify-start">
      <OrgAvatar name={org.name} />

      <div>
        <Text
          className={classNames(
            'text-md !font-medium leading-normal max-w-[150px] mb-1 break-all text-left !flex-nowrap'
          )}
          title={org.sandbox_mode ? 'Org is in sandbox mode' : undefined}
        >
          {org.sandbox_mode && <TestTube className="text-md" />}
          <span
            className={classNames('', {
              'max-w-[120px]': org.sandbox_mode,
              'truncate !inline': org.name.length >= 16,
            })}
          >
            {org.name}
          </span>
        </Text>
        <StatusBadge status={org.status} isWithoutBorder />
      </div>
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
        <Text variant="label">GitHub Connections</Text>
        <Link
          className="flex items-center gap-2 text-sm font-medium"
          href={`https://github.com/apps/${GITHUB_APP_NAME}/installations/new?state=${org.id}`}
        >
          <Plus className="text-lg" />
          Add
        </Link>
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
  return (
    <div className="flex flex-col gap-4">
      <Text className="px-4" variant="label">
        Organizations
      </Text>
      <nav className="flex flex-col gap-0 px-1">
        {orgs.map((org) => (
          <NextLink
            className="flex items-center justify-start gap-4 rounded-md p-2 hover:bg-cool-grey-600/20"
            key={org.id}
            href={`/${org.id}/apps`}
          >
            <OrgAvatar name={org.name} />
            <span>
              <Text
                className="break-all text-md font-medium leading-normal mb-1 !flex-nowrap"
                title={org.sandbox_mode ? 'Org is in sandbox mode' : undefined}
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
  initOrg: TOrg
  initOrgs: Array<TOrg>
}

export const OrgSwitcher: FC<IOrgSwitcher> = ({ initOrg, initOrgs }) => {
  useEffect(() => {
    async function setSession() {
      await setOrgSessionCookie(initOrg.id)
    }

    setSession()
  }, [])

  return (
    <Dropdown
      className="w-full"
      hasCustomPadding
      id="test"
      isFullWidth
      text={<OrgSummary org={initOrg} />}
      position="overlay"
      alignment="overlay"
    >
      <div className="flex flex-col gap-4 overflow-auto max-h-[500px] pb-2">
        <div className="pt-2 px-4">
          <OrgSummary org={initOrg} />
          <ClickToCopy className="mt-4">
            <Text variant="id">{initOrg.id}</Text>
          </ClickToCopy>
        </div>
        <OrgVCSConnectionsDetails org={initOrg} />
        <OrgsNav orgs={initOrgs} />
      </div>
    </Dropdown>
  )
}
