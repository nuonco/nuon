import { z } from 'zod'

export type ClaimCondition = { key: string; value: string }

export type OIDCPreset = 'github_actions' | 'custom'

export type OIDCTrustPolicyFormInput = {
  name: string
  issuerUrl: string
  audience: string
  role: string
  tokenDurationSeconds: string
  claimConditions: ClaimCondition[]
}

export const GITHUB_ACTIONS_ISSUER =
  'https://token.actions.githubusercontent.com'

export const hasSubCondition = (claimConditions: ClaimCondition[]) =>
  claimConditions.some(
    (condition) => condition.key.trim() === 'sub' && condition.value.trim()
  )

export const githubSubClaim = (repoFullName: string, branch: string) =>
  `repo:${repoFullName}:ref:refs/heads/${branch}`

export const defaultRepoPolicyName = (
  repoFullName: string,
  reservedNames: string[] = []
) => {
  const taken = new Set(
    reservedNames.map((reserved) => reserved.trim().toLowerCase())
  )
  const baseName = `github-${repoFullName.split('/').pop() ?? repoFullName}`
  let name = baseName
  for (let n = 2; taken.has(name.toLowerCase()); n++) {
    name = `${baseName}-${n}`
  }
  return name
}

export interface OIDCFormValues {
  name: string
  issuerUrl: string
  audience: string
  role: string
  tokenDurationSeconds: string
  claimConditions: ClaimCondition[]
}

export const buildCreateOIDCSchema = (reservedNames: string[] = []) =>
  z
    .object({
      name: z.string().trim().min(1, 'Name is required'),
      issuerUrl: z
        .string()
        .trim()
        .regex(/^https?:\/\/.+/i, 'Must be an absolute http or https URL'),
      audience: z.string().trim().min(1, 'Audience is required'),
      role: z.string(),
      tokenDurationSeconds: z.string(),
      claimConditions: z.array(
        z.object({ key: z.string(), value: z.string() })
      ),
    })
    .superRefine((v, ctx) => {
      const taken = reservedNames.some(
        (reserved) => reserved.trim().toLowerCase() === v.name.trim().toLowerCase()
      )
      if (taken) {
        ctx.addIssue({
          code: 'custom',
          path: ['name'],
          message: `A trust policy named ${v.name.trim()} already exists. Choose a different name.`,
        })
      }
    })
