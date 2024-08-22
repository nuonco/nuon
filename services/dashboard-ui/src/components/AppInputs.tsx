import React, { type FC } from 'react'
import { Text } from '@/components'
import type { TAppInputConfig } from '@/types'

export interface IAppInputConfig {
  inputConfig: TAppInputConfig
}

export const AppInputConfig: FC<IAppInputConfig> = ({ inputConfig }) => {
  return (
    <div className="flex flex-col gap-2">
      <div className="grid grid-cols-7 gap-4 px-3 py-4">
        <Text className="col-span-2" variant="label">
          Name
        </Text>
        <Text className="col-span-2" variant="label">
          Description
        </Text>
        <Text variant="label">Default value</Text>
        <Text variant="label">Required</Text>
        <Text variant="label">Sensitive</Text>
      </div>
      {inputConfig.input_groups.map((group, i) => (
        <div className="divide-y" key={`${group.id}-${i}`}>
          <div className="px-3 py-2 bg-gray-500/5 border-t grid grid-cols-7 gap-4">
            <Text className="col-span-2" variant="overline">
              {group.display_name}
            </Text>
            <Text className="col-span-2" variant="overline">
              {group.description}
            </Text>
          </div>

          <div className="divide-y">
            {group.app_inputs.map((input, ii) => (
              <div
                key={`${input.id}-${ii}`}
                className="grid grid-cols-7 gap-4 px-3 py-4"
              >
                <Text className="col-span-2 gap-4" variant="caption">
                  <span className="font-semibold">{input.display_name}</span>{' '}
                  <span className="font-mono font-[10px]">{input.name}</span>
                </Text>

                <Text className="col-span-2" variant="caption">
                  {input.description}
                </Text>

                <Text variant="caption">{input.default || 'No default'}</Text>

                <Text variant="caption">
                  {input.required ? 'True' : 'False'}
                </Text>

                <Text variant="caption">
                  {input.sensitive ? 'True' : 'False'}
                </Text>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
