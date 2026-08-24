import { z } from 'zod'

export const admissionRulesSchema = z.object({
  admissionPolicy: z.enum(['enqueue', 'strict']),
  scheduledOperationsPolicy: z.enum(['direct', 'enqueue']),
})

export type AdmissionRulesValues = z.infer<typeof admissionRulesSchema>

export const installQueueReorderSchema = z.object({
  queuePosition: z.number().int().min(1),
})

export type InstallQueueReorderValues = z.infer<
  typeof installQueueReorderSchema
>
