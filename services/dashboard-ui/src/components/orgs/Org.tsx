'use client'

import React, { type FC, useEffect, useState } from 'react'
import { FaGithub } from 'react-icons/fa'
import { Card, Heading, Link, PageHeader, Status, Text } from '@/components'
import { TOrg } from '@/types'

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

export const OrgPageHeader: FC<TOrg> = (org) => {
  return (
    <PageHeader
      info={<OrgStatus org={org} />}
      title={<OrgTitle {...org} />}
      summary={<OrgVCSConnections {...org} />}
    />
  )
}

export const OrgTitle: FC<TOrg> = ({ id, name }) => {
  return (
    <span className="flex flex-col gap-2">
      <Text variant="overline">{id}</Text>
      <Heading level={1} variant="title">
        {name}
      </Heading>
    </span>
  )
}

export const OrgVCSConnections: FC<TOrg> = ({ vcs_connections }) => {
  return (
    <Text className="flex flex-wrap gap-4 items-center" variant="status">
      {vcs_connections?.length &&
        vcs_connections?.map((vcs) => (
          <span key={vcs?.id} className="flex gap-1 items-center">
            <FaGithub className="text-md" /> {vcs?.github_install_id}
          </span>
        ))}
    </Text>
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
    pollStatus = setInterval(fetchStatus, 15000)
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
