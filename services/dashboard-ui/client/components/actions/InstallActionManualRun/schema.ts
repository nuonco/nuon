import { z } from 'zod'

export interface ManualRunValues {
  configVars: Record<string, string>
  customVars: Array<{ name: string; value: string }>
}

export const buildManualRunSchema = (configVarNames: string[]) =>
  z.object({
    configVars: z.object(
      Object.fromEntries(
        configVarNames.map((name) => [name, z.string().min(1, 'Required')])
      )
    ),
    customVars: z.array(
      z.object({
        name: z.string().min(1, 'Name is required'),
        value: z.string().min(1, 'Value is required'),
      })
    ),
  })
