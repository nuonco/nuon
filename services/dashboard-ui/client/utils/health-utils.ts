export const HEALTH_UNHEALTHY = 'unhealthy'
export const HEALTH_DEGRADED = 'degraded'
export const HEALTH_PROGRESSING = 'progressing'
export const HEALTH_HEALTHY = 'healthy'
export const HEALTH_UNKNOWN = 'unknown'
export const HEALTH_NOT_APPLICABLE = 'not-applicable'

// Mirrors componentHealthSeverity in the evaluator: unknown and not-applicable
// share zero severity because neither is a verdict — one is "tried and could
// not tell", the other "nothing here exposes health".
const HEALTH_SEVERITY: Record<string, number> = {
  [HEALTH_UNHEALTHY]: 4,
  [HEALTH_DEGRADED]: 3,
  [HEALTH_PROGRESSING]: 2,
  [HEALTH_HEALTHY]: 1,
}

export function healthSeverity(health?: string): number {
  return HEALTH_SEVERITY[health ?? ''] ?? 0
}

export function isFailingHealth(health?: string): boolean {
  return health === HEALTH_UNHEALTHY || health === HEALTH_DEGRADED
}

export function bearsHealthVerdict(health?: string): boolean {
  return healthSeverity(health) > 0
}

// Falls back the way the evaluator does when nothing was assessed: unknown
// outranks not-applicable, because "tried and could not tell" is more
// informative than "nothing here exposes health".
export function worstHealth(healths: (string | undefined)[]): string {
  const assessed = healths.filter(bearsHealthVerdict)
  if (assessed.length > 0) {
    return assessed.reduce<string>(
      (worst, health) =>
        healthSeverity(health) > healthSeverity(worst) ? health! : worst,
      assessed[0]!
    )
  }
  return healths.includes(HEALTH_NOT_APPLICABLE) &&
    !healths.includes(HEALTH_UNKNOWN)
    ? HEALTH_NOT_APPLICABLE
    : HEALTH_UNKNOWN
}

export function compareHealthSeverityDesc(a?: string, b?: string): number {
  return healthSeverity(b) - healthSeverity(a)
}
