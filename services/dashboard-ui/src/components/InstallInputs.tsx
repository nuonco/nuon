import React, { type FC } from 'react'
import { Text } from '@/components'
import type { TInstall } from '@/types'

export interface IInstallInputs {
  inputs: TInstall['install_inputs']
}

export const InstallInputs: FC<IInstallInputs> = ({ inputs }) => {
  return (
    <div className="flex flex-col gap-2">
      <div className="grid grid-cols-3 gap-4">
        <Text variant="label">Name</Text>
        <Text variant="label">Value</Text>
      </div>

      <div>
        {inputs.map((input, ii) => (
          <div className="divide-y" key={`${input.id}-${ii}`}>
            {input?.redacted_values
              ? Object.keys(input.redacted_values).map((key, i) => (
                  <div
                    key={`${key}-${i}`}
                    className="grid grid-cols-3 gap-4 py-2"
                  >
                    <Text variant="caption">{key}</Text>
                    <Text className="col-span-2 break-all" variant="caption">
                      {input.redacted_values[key]}
                    </Text>
                  </div>
                ))
              : input?.values &&
                Object.keys(input.values).map((key, i) => (
                  <div
                    key={`${key}-${i}`}
                    className="grid grid-cols-3 gap-4 py-2"
                  >
                    <Text variant="caption">{key}</Text>
                    <Text className="col-span-2 break-all" variant="caption">
                      {input.values[key]}
                    </Text>
                  </div>
                ))}
          </div>
        ))}
      </div>
    </div>
  )
}
