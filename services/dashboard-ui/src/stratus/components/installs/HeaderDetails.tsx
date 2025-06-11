'use client'

import React, { type FC } from 'react'
import { CheckCircle } from '@phosphor-icons/react/dist/ssr'
import { Link, Text } from '@/stratus/components/common'
import { HeaderDetails } from '@/stratus/components/dashboard'
import { DetailedStatus, Status } from '@/stratus/components/statuses'
import { useInstall } from '@/stratus/context'
import { InstallManageDropdown } from './ManageDropdown'

export const InstallHeaderDetails: FC = () => {
  const { install } = useInstall()
  const pathRoot = `/stratus/${install?.org_id}/installs/${install?.id}`

  return (
    <HeaderDetails>
      <div className="flex flex-col gap-1.5 self-start">
        <Text variant="subtext" theme="muted">
          App config
        </Text>
        <Text variant="subtext" weight="strong">
          <Link href={`/stratus/${install?.org_id}/apps/${install?.app?.id}`}>
            {install?.app?.name}
          </Link>
        </Text>
      </div>

      <DetailedStatus
        status={{ status: install?.runner_status }}
        tooltip={{
          tipContent: (
            <span className="flex flex-col gap-2 py-2 w-60">
              <Text
                className="flex gap-1 items-center"
                variant="subtext"
                weight="strong"
              >
                <CheckCircle
                  size="18"
                  className="text-green-600"
                  weight="bold"
                />
                Runner is provisioned
              </Text>
              <Text variant="label" theme="muted">
                {install.runner_status_description}
              </Text>
              <Text className="mt-2" variant="subtext" weight="strong">
                <Link href={`${pathRoot}/runner`}>View details</Link>
              </Text>
            </span>
          ),
          position: 'bottom',
        }}
        title="Runner"
      />

      <DetailedStatus
        status={{ status: install?.sandbox_status }}
        tooltip={{
          tipContent: (
            <span className="flex flex-col gap-2 py-2 w-60">
              <Text
                className="flex gap-1 items-center"
                variant="subtext"
                weight="strong"
              >
                <CheckCircle
                  size="18"
                  className="text-green-600"
                  weight="bold"
                />
                Sandbox is provisioned
              </Text>
              <Text variant="label" theme="muted">
                {install?.sandbox_status_description}
              </Text>
              <Text className="mt-2" variant="subtext" weight="strong">
                <Link
                  href={`${pathRoot}/sandbox/sandbox-runs/${install?.install_sandbox_runs?.at(0)?.id}`}
                >
                  View details
                </Link>
              </Text>
            </span>
          ),
          position: 'bottom',
        }}
        title="Sandbox"
      />

      <DetailedStatus
        status={{ status: install?.composite_component_status }}
        tooltip={{
          position: 'bottom',
          tipContent: (
            <span className="flex flex-col gap-2 py-2 w-60">
              <Text
                className="flex gap-1 items-center"
                variant="subtext"
                weight="strong"
              >
                <CheckCircle
                  size="18"
                  className="text-green-600"
                  weight="bold"
                />
                Components are deployed
              </Text>
              <Text variant="label" theme="muted">
                {install?.composite_component_status_description}
              </Text>
              <div className="flex flex-col gap-1 max-h-96 overflow-auto bg-white/10 bg-dark-grey-50/10 rounded p-2">
                {install?.install_components?.map((c) => (
                  <Text
                    className="flex items-center gap-1.5"
                    key={c?.id}
                    variant="subtext"
                    weight="strong"
                  >
                    <Status status={c?.status || 'unkown'} isWithoutText />
                    {c?.component?.name}
                  </Text>
                ))}
              </div>

              <Text className="mt-2" variant="subtext" weight="strong">
                <Link href={`${pathRoot}/components`}>View details</Link>
              </Text>
            </span>
          ),
        }}
        title="Components"
      />

      <InstallManageDropdown />
    </HeaderDetails>
  )
}
