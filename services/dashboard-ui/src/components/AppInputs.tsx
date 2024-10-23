import React, { type FC } from 'react'
import { Text } from '@/components/Typography'
import type { TAppInputConfig } from '@/types'

export interface IAppInputConfig {
  inputConfig: TAppInputConfig
}

export const AppInputConfig: FC<IAppInputConfig> = ({ inputConfig }) => {
  return inputConfig ? (
    <div className="flex flex-col gap-2">
      <div className="grid grid-cols-8 gap-4 px-3 py-2 text-cool-grey-600 dark:text-cool-grey-500 text-base">
        <Text className="!font-medium col-span-2">Name</Text>
        <Text className="!font-medium col-span-2">Description</Text>
        <Text className="!font-medium col-span-2">Default</Text>
        <Text className="!font-medium">Required</Text>
        <Text className="!font-medium">Sensitive</Text>
      </div>
      {inputConfig.input_groups.map((group, i) => (
        <div className="divide-y" key={`${group.id}-${i}`}>
          <div className="px-3 py-2 bg-cool-grey-50 dark:bg-dark-grey-200 text-cool-grey-600 dark:text-cool-grey-500 text-base border-t grid grid-cols-8 gap-4 items-start">
            <Text className="col-span-2 !font-medium">
              {group.display_name}
            </Text>
            <Text className="col-span-2 text-sm">{group.description}</Text>
          </div>

          <div className="divide-y">
            {group.app_inputs.map((input, ii) => (
              <div
                key={`${input.id}-${ii}`}
                className="grid grid-cols-8 gap-4 px-3 py-4 items-start"
              >
                <div className="col-span-2 gap-2 flex flex-col items-start justify-start">
                  <Text className="!font-medium">{input.display_name}</Text>{' '}
                  <Text className="font-mono text-cool-grey-600 dark:text-cool-grey-500 text-sm break-all !inline truncate max-w-[150px]">
                    {input.name}
                  </Text>
                </div>

                <Text className="col-span-2 text-sm">{input.description}</Text>

                <Text className="col-span-2 text-sm">
                  {input.default || 'No default'}
                </Text>

                <Text className="text-sm">{input.required ? 'True' : 'False'}</Text>

                <Text className="text-sm">{input.sensitive ? 'True' : 'False'}</Text>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  ) : <Text>No app inputs configured</Text>
}
