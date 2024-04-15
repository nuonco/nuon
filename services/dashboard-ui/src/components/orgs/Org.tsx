'use client'

import React, { type FC, useEffect, useState } from 'react'
import { FaAws, FaGitAlt, FaGithub } from 'react-icons/fa'
import { Card, Heading, Link, Status, Text } from '@/components'
import { TOrg } from '@/types'
import { Interval } from 'luxon'

export const OrgCard: FC<TOrg> = ({ status, status_description, id, name }) => {
  return (
    <Card>
      <OrgStatus org={{ id, status, status_description }} isCompact />
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

export const OrgHeading: FC<TOrg> = ({
  vcs_connections,
  id,
  name,
  status,
  status_description,
}) => {
  return (
    <div className="flex flex-auto gap-8 items-end border-b pb-6">
      <div className="flex flex-col flex-auto gap-2">
        <span>
          <Text variant="overline">{id}</Text>
          <Heading level={1} variant="title">
            {name}
          </Heading>
        </span>

        <Text className="flex flex-wrap gap-4 items-center" variant="status">
          {vcs_connections?.length &&
            vcs_connections?.map((vcs) => (
              <span key={vcs?.id} className="flex gap-1 items-center">
                <FaGithub className="text-md" /> {vcs?.github_install_id}
              </span>
            ))}
        </Text>
      </div>

      <div className="flex flex-col flex-auto gap-4">
        <OrgStatus org={{ id, status, status_description }} />
      </div>
    </div>
  )
}

export const OrgStatus: FC<{ org: TOrg; isCompact?: boolean }> = ({
  isCompact = false,
  org: { id, status, status_description },
}) => {
  const [{ orgHealth, orgStatus }, setStatus] = useState({
    orgStatus: { status, status_description },
    orgHealth: { status: 'waiting', status_description: 'Checking org health' },
  })
  const fetchStatus = () => {
    fetch(`/api/${id}/status`)
      .then((res) => res.json().then((s) => setStatus(s)))
      .catch(console.error)
  }

  useEffect(() => {
    fetchStatus()
  }, [])

  let pollStatus: NodeJS.Timeout
  useEffect(() => {
    pollStatus = setInterval(fetchStatus, 5000)
    return () => clearInterval(pollStatus)
  }, [orgHealth, orgStatus])

  return (
    <div className="flex flex-grow gap-6">
      <Status
        status={orgStatus?.status}
        description={!isCompact && orgStatus?.status_description}
        label={isCompact ? orgStatus?.status : 'Status'}
        isLabelStatusText={isCompact}
      />
      <Status
        status={orgHealth?.status}
        description={!isCompact && orgHealth?.status_description}
        label={isCompact ? orgHealth?.status : 'Health'}
        isLabelStatusText={isCompact}
      />
    </div>
  )
}
