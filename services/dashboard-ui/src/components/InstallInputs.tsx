// TODO(nnnat): remove once we have this API changes on prod
// @ts-nocheck
import React, { type FC } from 'react'
import { Text } from '@/components/Typography'
import type { TInstall } from '@/types'

export interface IInstallInputs {
  inputs: TInstall['install_inputs'] & {
    redacted_values: Array<Record<string, string>>
  }
}

export const InstallInputs: FC<IInstallInputs> = ({ inputs }) => {
  return (
    <div className="divide-y">
      <div className="grid grid-cols-3 gap-4 pb-3">
        <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
          Name
        </Text>
        <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
          Value
        </Text>
      </div>

      <div>
        {inputs.map((input, ii) => (
          <div className="divide-y" key={`${input.id}-${ii}`}>
            {input?.redacted_values
              ? Object.keys(input.redacted_values).map((key, i) => (
                  <div
                    key={`${key}-${i}`}
                    className="grid grid-cols-3 gap-4 py-3"
                  >
                    <Text className="font-mono text-sm break-all !inline truncate max-w-[200px]">{key}</Text>
                    <Text className="col-span-2 break-all text-sm !inline truncate max-w-[200px]">
                      {input.redacted_values[key]}
                    </Text>
                  </div>
                ))
              : input?.values &&
                Object.keys(input.values).map((key, i) => (
                  <div
                    key={`${key}-${i}`}
                    className="grid grid-cols-3 gap-4 py-3"
                  >
                    <Text className="font-mono text-sm !inline truncate max-w-[200px]">{key}</Text>
                    <Text className="col-span-2 break-all text-sm !inline truncate max-w-[200px]">
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
