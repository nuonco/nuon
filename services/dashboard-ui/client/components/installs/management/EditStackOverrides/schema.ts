import { z } from 'zod'

export const editStackOverridesSchema = z.object({
  vpcUrl: z.string(),
  runnerUrl: z.string(),
  customStacks: z.array(
    z.object({
      name: z.string(),
      template_url: z.string(),
      index: z.number(),
      parameters: z.record(z.string(), z.string()).optional(),
    })
  ),
})

export type EditStackOverridesValues = z.infer<typeof editStackOverridesSchema>
export type CustomNestedStackEntry =
  EditStackOverridesValues['customStacks'][number]
