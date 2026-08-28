import { z } from 'zod'
import type { TRunbookInput } from '@/lib/ctl-api/apps/runbooks'

export interface RunbookFormValues {
  inputs: Record<string, string | boolean>
}

export const isBooleanInput = (input: TRunbookInput) =>
  input.type === 'bool' || input.default === 'true' || input.default === 'false'

export const buildRunbookSchema = (inputs: TRunbookInput[]) =>
  z.object({
    inputs: z.object(
      Object.fromEntries(
        inputs.map((input) => {
          const label = input.display_name || input.name
          if (isBooleanInput(input)) return [input.name, z.boolean()]
          const base = z.string()
          return [
            input.name,
            input.required ? base.min(1, `${label} is required`) : base,
          ]
        })
      )
    ),
  })
