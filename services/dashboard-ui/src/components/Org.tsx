'use client'

import React, { type FC, useCallback, useEffect, useState } from 'react'
import { FaGithub } from 'react-icons/fa'
import { Card, Heading, Link, Status, Text } from '@/components'
import { useOrgContext } from '@/context'
import { GITHUB_APP_NAME, SHORT_POLL_DURATION } from '@/utils'

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
          <span key={vcs?.id} className="flex gap-1 items-center">
            <FaGithub className="text-md" /> {vcs?.github_install_id}
          </span>
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
