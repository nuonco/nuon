import React, { type FC } from 'react'
import { FaAws } from 'react-icons/fa'
import { VscAzure } from 'react-icons/vsc'
import { Text } from '@/components'
import type { TAppRunnerConfig } from '@/types'

export interface IAppRunnerConfig {
  runnerConfig: TAppRunnerConfig
}

export const AppRunnerConfig: FC<IAppRunnerConfig> = ({ runnerConfig }) => {
  return (
    <div className="flex gap-4">
      <span className="flex flex-col gap-2">
        <Text variant="overline">Platform:</Text>
        <Text variant="caption">
          <Platform platform={runnerConfig.cloud_platform} />
        </Text>
      </span>

      <span className="flex flex-col gap-2">
        <Text variant="overline">Runner type:</Text>
        <Text variant="caption">{runnerConfig.app_runner_type}</Text>
      </span>
    </div>
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
