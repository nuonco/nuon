import { z } from 'zod'

export const runAdhocActionSchema = z
  .object({
    name: z.string().max(255),
    inputMode: z.enum(['command', 'script']),
    command: z.string(),
    inlineContents: z.string(),
    timeout: z.string(),
    envVars: z.array(
      z.object({
        name: z.string(),
        value: z.string(),
      })
    ),
  })
  .superRefine((val, ctx) => {
    if (val.inputMode === 'command' && !val.command.trim()) {
      ctx.addIssue({
        code: 'custom',
        path: ['command'],
        message: 'Command is required',
      })
    }
    if (val.inputMode === 'script' && !val.inlineContents.trim()) {
      ctx.addIssue({
        code: 'custom',
        path: ['inlineContents'],
        message: 'Script is required',
      })
    }
    if (val.timeout.trim() !== '') {
      const n = Number(val.timeout)
      if (!Number.isInteger(n) || n < 1 || n > 3600) {
        ctx.addIssue({
          code: 'custom',
          path: ['timeout'],
          message: 'Timeout must be between 1 and 3600 seconds',
        })
      }
    }
    val.envVars.forEach((ev, i) => {
      if (!ev.name.trim()) {
        ctx.addIssue({
          code: 'custom',
          path: ['envVars', i, 'name'],
          message: 'Name is required',
        })
      }
      if (!ev.value.trim()) {
        ctx.addIssue({
          code: 'custom',
          path: ['envVars', i, 'value'],
          message: 'Value is required',
        })
      }
    })
  })

export type RunAdhocActionValues = z.infer<typeof runAdhocActionSchema>
