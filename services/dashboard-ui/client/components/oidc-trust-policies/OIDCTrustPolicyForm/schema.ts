import { z } from 'zod'

export type OIDCTrustPolicyMode = 'create' | 'edit'

export type ClaimCondition = { key: string; value: string }

export type OIDCPreset = 'github_actions' | 'custom'

export const GITHUB_ACTIONS_ISSUER =
  'https://token.actions.githubusercontent.com'

export const hasSubCondition = (claimConditions: ClaimCondition[]) =>
  claimConditions.some(
    (condition) => condition.key.trim() === 'sub' && condition.value.trim()
  )

export type GithubRepoSubjectIds = {
  ownerId: number
  repoId: number
}

// GitHub can issue either a legacy `repo:{owner}/{repo}:ref:...` sub claim or
// an immutable one with numeric IDs appended (`{owner}@{ownerId}`). The
// `{,@id}` alternation matches either form without weakening the ID check.
export const githubSubClaim = (
  repoFullName: string,
  branch: string,
  ids?: GithubRepoSubjectIds
) => {
  if (!ids) {
    return `repo:${repoFullName}:ref:refs/heads/${branch}`
  }
  const [owner, repo] = repoFullName.split('/')
  return `repo:${owner}{,@${ids.ownerId}}/${repo}{,@${ids.repoId}}:ref:refs/heads/${branch}`
}

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
  enabled: boolean
  claimConditions: ClaimCondition[]
}

export const buildOIDCSchema = ({
  mode,
  reservedNames = [],
}: {
  mode: OIDCTrustPolicyMode
  reservedNames?: string[]
}) =>
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
      enabled: z.boolean(),
      claimConditions: z.array(
        z.object({ key: z.string(), value: z.string() })
      ),
    })
    .superRefine((v, ctx) => {
      if (mode !== 'create') return
      const taken = reservedNames.some(
        (reserved) =>
          reserved.trim().toLowerCase() === v.name.trim().toLowerCase()
      )
      if (taken) {
        ctx.addIssue({
          code: 'custom',
          path: ['name'],
          message: `A trust policy named ${v.name.trim()} already exists. Choose a different name.`,
        })
      }
    })
