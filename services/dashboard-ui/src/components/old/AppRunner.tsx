import React, { type FC } from 'react'
import { FaAws } from 'react-icons/fa'
import { VscAzure } from 'react-icons/vsc'
import { Config, ConfigContent } from '@/components/old/Config'
import { Text } from '@/components/old/Typography'
import type { TAppRunnerConfig } from '@/types'

export interface IAppRunnerConfig {
  runnerConfig: TAppRunnerConfig
}

export const AppRunnerConfig: FC<IAppRunnerConfig> = ({ runnerConfig }) => {
  return runnerConfig ? (
    <Config>
      <ConfigContent
        label="Platform"
        value={<Platform platform={runnerConfig.cloud_platform} />}
      />
      <ConfigContent label="Runner type" value={runnerConfig.app_runner_type} />
    </Config>
  ) : (
    <Text>Missing app runner configuration</Text>
  )
}

export interface IPlatform {
  isIconOnly?: boolean
  platform: 'aws' | 'azure' | unknown
}

export const Platform: FC<IPlatform> = ({ isIconOnly = false, platform }) => {
  return (
    <span className="flex items-center gap-2">
      {platform === 'azure' ? (
        <>
          <VscAzure className="text-md" /> {!isIconOnly && 'Azure'}
        </>
      ) : (
        <>
          <FaAws className="text-xl mb-[-4px]" /> {!isIconOnly && 'Amazon'}
        </>
      )}
    </span>
  )
}
