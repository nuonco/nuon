'use client'

import React, { type FC, useState } from 'react'
import { useRouter } from 'next/navigation'
import { createAppInstall } from '@/actions/apps/create-app-install'
import { InstallForm } from '@/components/InstallForm'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useQuery } from '@/hooks/use-query'
import type { TOrg, TApp, TAppConfig } from '@/types'
import { OrgProvider } from '@/providers/org-provider'

interface InstallCreationStepContentProps {
  stepComplete: boolean
  onClose: () => void
  installId?: string
  appId?: string
  orgId?: string
}

export const InstallCreationStepContent: FC<
  InstallCreationStepContentProps
> = ({ stepComplete, onClose, installId, appId, orgId }) => {
  const router = useRouter()

  // Load org data
  const {
    data: org,
    isLoading: orgLoading,
    error: orgError,
  } = useQuery<TApp>({
    path: `/api/orgs/${orgId}`,
    enabled: !!orgId,
  })

  // Load app data
  const {
    data: app,
    isLoading: appLoading,
    error: appError,
  } = useQuery<TApp>({
    path: `/api/orgs/${orgId}/apps/${appId}`,
    enabled: !!appId && !!orgId,
  })

  // Load app configs
  const {
    data: configs,
    isLoading: configsLoading,
    error: configsError,
  } = useQuery<TAppConfig[]>({
    path: `/api/orgs/${orgId}/apps/${appId}/configs`,
    enabled: !!appId && !!orgId,
  })

  // Load latest app config
  const configId = configs?.at(0)?.id
  const {
    data: config,
    isLoading: configLoading,
    error: configError,
  } = useQuery<TAppConfig>({
    path: `/api/orgs/${orgId}/apps/${appId}/configs/${configId}?recurse=true`,
    enabled: !!appId && !!orgId && !!configId,
  })

  const isLoading = orgLoading || appLoading || configsLoading || configLoading
  const error = orgError || appError || configsError || configError

  if (stepComplete) {
    return (
      <div className="space-y-6">
        {/* Success Message - Shown when step is complete */}
        <div className="space-y-3 pb-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full" />
            <Text
              variant="semi-14"
              className="text-green-800 dark:text-green-200"
            >
              Install created successfully!
            </Text>
          </div>
          {installId && (
            <div className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg">
              <Text
                variant="reg-12"
                className="text-gray-600 dark:text-gray-400"
              >
                Install ID:{' '}
                <code className="font-mono text-gray-800 dark:text-gray-200">
                  {installId}
                </code>
              </Text>
            </div>
          )}
        </div>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="p-6">
        <Loading loadingText="Loading app configuration..." variant="stack" />
      </div>
    )
  }

  if (error?.error) {
    return (
      <div className="p-6">
        <Notice>{error?.error || 'Unable to load app configuration'}</Notice>
      </div>
    )
  }

  if (!app || !config) {
    return (
      <div className="p-6">
        <Notice>App configuration not found</Notice>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Instructions */}
      <Text>You are almost ready to create an install!</Text>
      <Text>
        Before creating the install, log into the AWS account you want to
        install into. Make sure you have the required permissions to create
        resources, and have not reached any quota limits.
      </Text>
      <Text>Complete the form below to create your install.</Text>

      {/* Embedded Install Form */}
      <div className="border border-gray-200 dark:border-gray-700 rounded-lg">
        <OrgProvider initOrg={org}>
          <InstallForm
            onSubmit={(formData) => {
              return createAppInstall({
                appId: app?.id,
                orgId: orgId,
                formData,
                path: window.location.pathname,
              })
            }}
            onSuccess={({ data: install, error, headers, status }) => {
              if (!error && status === 201) {
                router.push(
                  `/${orgId}/installs/${install?.id}/workflows/${headers?.['x-nuon-install-workflow-id']}?onboardingComplete=true`
                )
              }
            }}
            onCancel={() => {
              // Don't close the onboarding - just refresh to stay in flow
            }}
            platform={app?.runner_config.app_runner_type}
            inputConfig={{
              ...config.input,
              input_groups: nestInputsUnderGroups(
                config.input?.input_groups,
                config.input?.inputs
              ),
            }}
          />
        </OrgProvider>
      </div>
    </div>
  )
}

function nestInputsUnderGroups(
  groups: TAppConfig['input']['input_groups'],
  inputs: TAppConfig['input']['inputs']
) {
  return groups.map((group) => ({
    ...group,
    app_inputs: inputs.filter((input) => input.group_id === group.id),
  }))
}
