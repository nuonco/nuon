// @ts-nocheck
// TODO(nnnat): URLSearchParams typing is terrible.
// What we're doing now is legit but TS doesn't think so.
'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import {
  ArrowsClockwise,
  CaretRight,
  Heartbeat,
  Timer,
} from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Config, ConfigContent } from '@/components/Config'
import { Expand } from '@/components/Expand'
import { Grid } from '@/components/Grid'
import { Link } from '@/components/Link'
import { Loading, SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { useOrg } from '@/components/Orgs'
import { jobHrefPath, jobName } from '@/components/Runners/helpers'
import { StatusBadge } from '@/components/Status'
import { Time, Duration } from '@/components/Time'
import { Text } from '@/components/Typography'
import { restartOrgRunners } from '@/components/admin-actions'
import type { TRunner, TRunnerHeartbeat, TRunnerJob, TInstall } from '@/types'

export const AdminRunnerModal: FC = ({}) => {
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [installs, setInstalls] = useState<Array<TInstall>>()
  const [error, setError] = useState<string>()

  const fetchInstalls = () => {
    fetch(`/api/${org.id}/installs`)
      .then((res) =>
        res.json().then((ins) => {
          setInstalls(ins)
          setIsLoading(false)
        })
      )
      .catch((err) => {
        console.error(err?.message)
        setIsLoading(false)
        setError('Unable to load org installs')
      })
  }

  useEffect(() => {
    fetchInstalls()
  }, [])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className=""
              actions={
                <RestartRunnersButton
                  onSuccess={() => {
                    fetchInstalls()
                  }}
                />
              }
              heading={`All ${org.name} runners`}
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col divide-y">
                <div>
                  <Text className="mb-3" variant="med-14">
                    Org Runners
                  </Text>
                  <Grid variant="3-cols">
                    {org?.runner_group?.runners?.map((runner) => (
                      <GridCard key={runner.id}>
                        <RunnerCard
                          runner={runner}
                          href={`/${org.id}/runner`}
                        />
                      </GridCard>
                    ))}
                  </Grid>
                </div>

                <div className="pt-3 mt-6">
                  <Text className="mb-3" variant="med-14">
                    Install Runners
                  </Text>
                  {error ? (
                    <Notice>{error}</Notice>
                  ) : isLoading ? (
                    <Loading loadingText="Loading install runners" />
                  ) : installs && installs.length ? (
                    <Grid variant="3-cols">
                      {installs.map((install) => (
                        <GridCard key={install.id}>
                          <Text variant="med-12">{install.name} runner</Text>
                          <LoadRunnerCard
                            runnerId={install?.runner_id}
                            installId={install.id}
                          />
                        </GridCard>
                      ))}
                    </Grid>
                  ) : (
                    <Text>No installs</Text>
                  )}
                </div>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <div className="flex flex-col gap-2">
        <Text variant="reg-14">Manage all runners in this org</Text>
        <Button
          className="text-base"
          onClick={() => {
            setIsOpen(true)
          }}
        >
          Manage all runners
        </Button>
      </div>
    </>
  )
}

const GridCard: FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <div className="rounded border px-3 py-2 flex flex-col gap-3">
      {children}
    </div>
  )
}

const RunnerCard: FC<{ runner: TRunner; href: string }> = ({
  runner,
  href,
}) => {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between w-full">
        <Text variant="med-12" className="gap-2">
          <span className="animate-pulse">
            <StatusBadge
              status={runner?.status}
              isStatusTextHidden
              isWithoutBorder
            />
          </span>
          <span>{runner?.display_name}</span>
        </Text>
        <Link className="text-sm" href={href}>
          Details <CaretRight />
        </Link>
      </div>
      <Expand
        parentClass="border rounded"
        headerClass="px-3 py-2"
        id={runner.id}
        heading={<LoadRunnerHeartbeat runnerId={runner.id} />}
        expandContent={
          <div className="px-3 flex flex-col border-t divide-y">
            <div className="py-3 flex flex-col gap-3">
              <Text variant="med-12">Last shut-down job</Text>
              <LoadRunnerJob runnerId={runner.id} groups={['operations']} />
            </div>
            <div className="py-3 flex flex-col gap-3">
              <Text variant="med-12">Recent job</Text>
              <LoadRunnerJob
                runnerId={runner.id}
                statuses={[
                  'finished',
                  'error',
                  'timed-out',
                  'cancelled',
                  'not-attempted',
                ]}
              />
            </div>
          </div>
        }
      />
    </div>
  )
}

const LoadRunnerCard: FC<{ runnerId: string; installId: string }> = ({
  runnerId,
  installId,
}) => {
  const { org } = useOrg()
  const [isLoading, setIsLoading] = useState(true)
  const [runner, setRunner] = useState<TRunner>()
  const [error, setError] = useState<string>()

  useEffect(() => {
    fetch(`/api/${org.id}/runner/${runnerId}`)
      .then((res) =>
        res.json().then((rnr) => {
          setRunner(rnr)
          setIsLoading(false)
        })
      )
      .catch((err) => {
        console.error(err?.message)
        setIsLoading(false)
        setError('Unable to load install runner')
      })
  }, [])

  return error ? (
    <Notice>{error}</Notice>
  ) : isLoading ? (
    <Loading loadingText={`Loading ${runnerId} runner...`} />
  ) : (
    <RunnerCard
      runner={runner}
      href={`/${org.id}/installs/${installId}/runner-group/${runnerId}`}
    />
  )
}

const LoadRunnerHeartbeat: FC<{ runnerId: string }> = ({ runnerId }) => {
  const { org } = useOrg()
  const [isLoading, setIsLoading] = useState(true)
  const [heartbeat, setHeartbeat] = useState<TRunnerHeartbeat>()
  const [error, setError] = useState<string>()

  const fetchHeartbeat = () => {
    fetch(`/api/${org.id}/runner/${runnerId}/latest-heart-beat`)
      .then((res) =>
        res.json().then((rnr) => {
          setHeartbeat(rnr)
          setIsLoading(false)
        })
      )
      .catch((err) => {
        console.error(err?.message)
        setIsLoading(false)
        setError('Unable to load install runner')
      })
  }

  useEffect(() => {
    fetchHeartbeat()
  }, [])

  useEffect(() => {
    fetchHeartbeat()
  }, [org])

  return error ? (
    <Notice>{error}</Notice>
  ) : isLoading ? (
    <Loading loadingText={`Loading last hearbeat...`} />
  ) : (
    <div className="flex items-start gap-4">
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Version
        </Text>
        <Text variant="med-12">{heartbeat?.version}</Text>
      </span>
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Alive time
        </Text>
        <Text>
          <Timer size={14} />
          <Duration nanoseconds={heartbeat?.alive_time} variant="med-12" />
        </Text>
      </span>
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Last heartbeat seen
        </Text>
        <Text>
          <Heartbeat size={14} />
          <Time
            time={heartbeat?.created_at}
            format="relative"
            variant="med-12"
          />
        </Text>
      </span>
    </div>
  )
}

const LoadRunnerJob: FC<{
  runnerId: string
  groups?: Array<'operations'>
  statuses?: Array<
    'finished' | 'error' | 'timed-out' | 'not-attempted' | 'cancelled'
  >
}> = ({ runnerId, groups, statuses }) => {
  const { org } = useOrg()
  const [isLoading, setIsLoading] = useState(true)
  const [job, setJob] = useState<TRunnerJob>()
  const [error, setError] = useState<string>()

  const params = new URLSearchParams({
    limit: '1',
    ...(groups ? { groups } : {}),
    ...(statuses ? { statuses } : {}),
  }).toString()

  const fetchRecentJob = () => {
    fetch(
      `/api/${org.id}/runner/${runnerId}/jobs${params ? '?' + params : params}`
    )
      .then((res) =>
        res.json().then((jbs) => {
          setJob(jbs?.[0])
          setIsLoading(false)
        })
      )
      .catch((err) => {
        console.error(err?.message)
        setIsLoading(false)
        setError('Unable to load install runner')
      })
  }

  useEffect(() => {
    fetchRecentJob()
  }, [])

  useEffect(() => {
    fetchRecentJob()
  }, [org])

  return error ? (
    <Notice>{error}</Notice>
  ) : isLoading ? (
    <Loading loadingText={`Loading latest job...`} />
  ) : job ? (
    <div className="flex items-start justify-between">
      <Config className="">
        <ConfigContent
          label="Job"
          value={
            <span className="flex items-center gap-2">
              <StatusBadge
                status={job?.status}
                isWithoutBorder
                isStatusTextHidden
              />
              {jobName(job) || 'Unknown'}
            </span>
          }
        />

        <ConfigContent
          label="Updated at"
          value={<Time time={job?.updated_at} />}
        />
      </Config>
      {jobHrefPath(job) !== '' ? (
        <Link className="text-sm" href={`/${org.id}/${jobHrefPath(job)}`}>
          Details <CaretRight />
        </Link>
      ) : null}
    </div>
  ) : (
    <Text>No job to show.</Text>
  )
}

const RestartRunnersButton: FC<{ onSuccess: () => void }> = (props) => {
  const { org } = useOrg()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()

  return error ? (
    <Notice>{error}</Notice>
  ) : (
    <Button
      onClick={() => {
        setIsLoading(true)
        restartOrgRunners(org.id)
          .then((res) => {
            setIsLoading(false)
            if (res.status === 201 || res.status === 200) {
              props.onSuccess()
            } else {
              setError(
                'Unable to restart org runners, refresh page and try again.'
              )
            }
          })
          .catch((err) => {
            console.error(err?.message)
            setIsLoading(false)
            setError(
              'Unable to restart org runners, refresh page and try again.'
            )
          })
      }}
      className="text-base flex items-center gap-2"
      disabled={isLoading}
    >
      {isLoading ? (
        <>
          <SpinnerSVG /> Restarting runners
        </>
      ) : (
        <>
          <ArrowsClockwise size="16" /> Restart all runners
        </>
      )}
    </Button>
  )
}
