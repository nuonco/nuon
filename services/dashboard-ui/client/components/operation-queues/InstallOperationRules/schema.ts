import { z } from 'zod'

export const operationWindowCadences = ['anytime', 'weekly', 'monthly'] as const
export const outsideWindowPolicies = ['reject', 'approval', 'queue'] as const
export const operationRuleIds = [
  'actions',
  'runbooks',
  'sandbox-updates',
  'deployments',
  'break-glass',
] as const

export const installOperationRuleSchema = z.object({
  id: z.enum(operationRuleIds),
  enabled: z.boolean(),
  cadence: z.enum(operationWindowCadences),
  daysOfWeek: z.array(z.string()),
  dayOfMonth: z.string(),
  startTime: z.string(),
  endTime: z.string(),
  enforceOutsideWindow: z.boolean(),
  outsideWindowPolicy: z.enum(outsideWindowPolicies),
})

export const installOperationRulesSchema = z.object({
  timezone: z.string().min(1, 'Select a timezone'),
  rules: z.array(installOperationRuleSchema),
})

export type TOperationWindowCadence = (typeof operationWindowCadences)[number]
export type TOutsideWindowPolicy = (typeof outsideWindowPolicies)[number]
export type TOperationRuleId = (typeof operationRuleIds)[number]
export type InstallOperationRuleValues = z.infer<
  typeof installOperationRuleSchema
>
export type InstallOperationRulesValues = z.infer<
  typeof installOperationRulesSchema
>

export const makeOperationRule = (
  id: TOperationRuleId,
  overrides: Partial<InstallOperationRuleValues> = {}
): InstallOperationRuleValues => ({
  id,
  enabled: false,
  cadence: 'anytime',
  daysOfWeek: [],
  dayOfMonth: '1',
  startTime: '09:00',
  endTime: '17:00',
  enforceOutsideWindow: true,
  outsideWindowPolicy: 'queue',
  ...overrides,
})

export const makeOperationRules = (
  overrides: Partial<
    Record<TOperationRuleId, Partial<InstallOperationRuleValues>>
  > = {}
): InstallOperationRuleValues[] =>
  operationRuleIds.map((id) => makeOperationRule(id, overrides[id]))

export const isRuleWindowValid = (rule: InstallOperationRuleValues) => {
  if (!rule.enabled || rule.cadence === 'anytime') return true
  if (!rule.startTime || !rule.endTime || rule.startTime === rule.endTime)
    return false
  if (rule.cadence === 'weekly') return rule.daysOfWeek.length > 0
  return !!rule.dayOfMonth
}
