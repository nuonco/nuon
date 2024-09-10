'use client'

import React, { type FC, useCallback, useEffect, useState } from 'react'
import { FaGithub } from 'react-icons/fa'
import { GoPlus } from 'react-icons/go'
import { Card, Dropdown, Heading, Link, Status, Text } from '@/components'
import { OrgProvider, useOrgContext } from '@/context'
import type { TOrg } from '@/types'
import {
  GITHUB_APP_NAME,
  SHORT_POLL_DURATION,
  initialsFromString,
} from '@/utils'

export const OrgCard: FC = () => {
  const {
    org: { id, name },
  } = useOrgContext()

  return (
    <Card>
      <OrgStatus isCompact />
      <span>
        <Text variant="overline">{id}</Text>
        <Heading variant="subheading">{name}</Heading>
      </span>

      <Text variant="caption">
        <Link href={`/dashboard/${id}`}>Details</Link>
      </Text>
    </Card>
  )
}

export const OrgVCSConnections: FC = () => {
  const {
    org: { vcs_connections },
  } = useOrgContext()

  return (
    <>
      {vcs_connections?.length &&
        vcs_connections?.map((vcs) => (
          <Text
            key={vcs?.id}
            className="flex gap-1 items-center"
            variant="caption"
          >
            <FaGithub className="text-md" /> {vcs?.github_install_id}
          </Text>
        ))}
    </>
  )
}

export const OrgConnectGithubLink: FC = () => {
  const { org } = useOrgContext()

  return (
    <Link
      className="flex items-center gap-1"
      href={`https://github.com/apps/${GITHUB_APP_NAME}/installations/new?state=${org.id}`}
    >
      <FaGithub className="text-md" />
      Connect GitHub
    </Link>
  )
}

export const OrgStatus: FC<{ isCompact?: boolean }> = ({
  isCompact = false,
}) => {
  const {
    org: { status, status_description },
  } = useOrgContext()

  return (
    <div className="flex flex-grow gap-6">
      <Status
        status={status}
        description={!isCompact && status_description}
        label={isCompact ? status : 'Status'}
        isLabelStatusText={isCompact}
      />
      <OrgHealthStatus
        initStatus={{
          status: 'waiting',
          status_description: 'Checking org health',
        }}
        shouldPoll
        isCompact={isCompact}
      />
    </div>
  )
}

export const OrgHealthStatus: FC<{
  initStatus: Record<'status' | 'status_description', string>
  shouldPoll?: boolean
  isCompact?: boolean
}> = ({ initStatus, shouldPoll = false, isCompact = false }) => {
  const { org } = useOrgContext()
  const [health, setHealth] = useState(initStatus)

  const fetchOrgHealth = useCallback(() => {
    fetch(`/api/${org.id}/health`)
      .then((res) => res.json().then((e) => setHealth(e)))
      .catch(console.error)
  }, [org])

  useEffect(() => {
    fetchOrgHealth()
  }, [])

  useEffect(() => {
    if (shouldPoll) {
      const pollOrgHealth = setInterval(fetchOrgHealth, SHORT_POLL_DURATION)
      return () => clearInterval(pollOrgHealth)
    }
  }, [health, org, shouldPoll, fetchOrgHealth])

  return (
    <Status
      status={health.status}
      description={!isCompact && health?.status_description}
      label={isCompact ? health?.status : 'Health'}
      isLabelStatusText={isCompact}
    />
  )
}

export const CreateOrgForm: FC<{ action: any }> = ({ action }) => {
  return (
    <form className="flex flex-col gap-4 max-w-md" action={action}>
      <label className="flex flex-col flex-auto gap-2">
        <span className="font-semibold">Organization name</span>
        <input
          className="border bg-inherit rounded px-4 py-1.5 shadow-inner"
          name="name"
          type="text"
          required
        />
      </label>

      <button className="rounded text-sm text-gray-50 bg-fuchsia-600 hover:bg-fuchsia-700 focus:bg-fuchsia-700 active:bg-fuchsia-800 px-4 py-1.5 w-fit">
        Create organization
      </button>
    </form>
  )
}

// NOTE(nnnnat): new semiflat designed org switcher parts
export interface IOrgSummary {}

export const OrgSummary: FC<IOrgSummary> = () => {
  const {
    org: { name },
  } = useOrgContext()

  return (
    <div className="flex gap-4 items-center justify-start">
      <span className="flex items-center justify-center p-2 rounded bg-gray-600/20 w-[40px] h-[40px] font-light font-sans">
        {initialsFromString(name)}
      </span>

      <div>
        <Text variant="label">{name}</Text>
        <OrgStatus isCompact />
      </div>
    </div>
  )
}

export const OrgVCSConnectionsDetails: FC = () => {
  const {
    org: { id },
  } = useOrgContext()
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <Text variant="label">GitHub Connections</Text>
        <Link
          className="flex items-center gap-1 text-sm"
          href={`https://github.com/apps/${GITHUB_APP_NAME}/installations/new?state=${id}`}
        >
          <GoPlus className="text-md" />
          Add new
        </Link>
      </div>

      <div>
        <OrgVCSConnections />
      </div>
    </div>
  )
}

export interface IOrgsNav {
  orgs: Array<TOrg>
}

export const OrgsNav: FC<IOrgsNav> = ({ orgs }) => {
  return (
    <div className="border-t border-dotted flex flex-col gap-4 pt-4">
      <Text variant="label">Organizations</Text>
      <nav className="flex flex-col gap-2">
        {orgs.map((org) => (
          <Link
            className="flex items-center justify-start gap-2"
            key={org.id}
            href={`/beta/${org.id}/apps`}
          >
            <span className="flex items-center justify-center p-2 rounded bg-gray-600/20 w-[30px] h-[30px] font-sans font-light text-sm text-gray-950 dark:text-gray-50">
              {initialsFromString(org.name)}
            </span>
            <Text>{org.name}</Text>
          </Link>
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
  return (
    <OrgProvider initOrg={initOrg}>
      <Dropdown
        className="w-full"
        id="test"
        text={<OrgSummary />}
        position="beside"
        alignment="right"
      >
        <div className="flex flex-col gap-4 p-4 w-72 overflow-auto max-h-[500px]">
          <OrgSummary />
          <OrgVCSConnectionsDetails />
          <OrgsNav orgs={initOrgs} />
        </div>
      </Dropdown>
    </OrgProvider>
  )
}
