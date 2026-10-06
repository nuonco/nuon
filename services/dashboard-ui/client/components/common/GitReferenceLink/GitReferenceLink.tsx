import type { TIconVariant } from '@/components/common/Icon'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text, type TTextVariant, type TTextWeight } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { externalGitUrl, gitReferenceUrl } from '@/utils/vcs-urls'

const COMMIT_LENGTH = 7

type TGitReference =
  | { type: 'commit'; sha: string; repo?: string; href?: string }
  | { type: 'pull-request'; number: number; repo?: string; href?: string; label?: string }
  | { type: 'tag'; tag: string; repo?: string; href?: string }

export interface IGitReferenceLink {
  reference: TGitReference
  showIcon?: boolean
  className?: string
  textVariant?: TTextVariant
  weight?: TTextWeight
  variant?: 'default' | 'inline'
}

const iconFor = (reference: TGitReference): TIconVariant => {
  if (reference.type === 'pull-request') return 'GitPullRequestIcon'
  if (reference.type === 'tag') return 'TagIcon'
  return 'GitCommitIcon'
}

const labelFor = (reference: TGitReference) => {
  if (reference.type === 'commit') return reference.sha.slice(0, COMMIT_LENGTH)
  if (reference.type === 'pull-request') {
    return reference.label ?? `PR #${reference.number}`
  }
  return reference.tag
}

const accessibleName = (reference: TGitReference) => {
  if (reference.type === 'commit') return `Commit ${reference.sha}`
  if (reference.type === 'pull-request') return `Pull request ${reference.number}`
  return `Tag ${reference.tag}`
}

const hrefFor = (reference: TGitReference) => {
  const explicit = externalGitUrl(reference.href)
  if (explicit) return explicit
  if (reference.type === 'commit') {
    return gitReferenceUrl(reference.repo, { type: 'commit', sha: reference.sha })
  }
  if (reference.type === 'pull-request') {
    return gitReferenceUrl(reference.repo, {
      type: 'pull-request',
      number: reference.number,
    })
  }
  return gitReferenceUrl(reference.repo, { type: 'tag', tag: reference.tag })
}

export const GitReferenceLink = ({
  reference,
  showIcon = true,
  className,
  textVariant = 'subtext',
  weight = 'normal',
  variant = 'default',
}: IGitReferenceLink) => {
  if (reference.type === 'commit' && !reference.sha) return null
  if (reference.type === 'pull-request' && !reference.number) return null
  if (reference.type === 'tag' && !reference.tag) return null

  const href = hrefFor(reference)
  const label = labelFor(reference)
  const name = accessibleName(reference)
  const mono = reference.type !== 'pull-request'
  const content = (
    <>
      {showIcon ? <Icon variant={iconFor(reference)} size="1em" /> : null}
      {label}
    </>
  )

  if (!href) {
    return (
      <Text
        as="span"
        variant={textVariant}
        weight={weight}
        theme="neutral"
        family={mono ? 'mono' : 'sans'}
        flex={showIcon}
        className={className}
        title={reference.type === 'commit' ? reference.sha : undefined}
        aria-label={name}
      >
        {content}
      </Text>
    )
  }

  return (
    <Link
      href={href}
      isExternal
      showExternalIcon={false}
      variant={variant}
      textVariant={textVariant}
      className={cn(
        weight === 'strong' && 'font-strong',
        weight === 'stronger' && 'font-stronger',
        mono && 'font-mono',
        className
      )}
      title={reference.type === 'commit' ? reference.sha : undefined}
      aria-label={name}
    >
      {content}
    </Link>
  )
}

export interface ICommitLink {
  sha?: string
  repo?: string
  href?: string
  showIcon?: boolean
  className?: string
  textVariant?: TTextVariant
  weight?: TTextWeight
  variant?: 'default' | 'inline'
}

export const CommitLink = ({
  sha,
  repo,
  href,
  ...props
}: ICommitLink) => {
  if (!sha) return null
  return <GitReferenceLink reference={{ type: 'commit', sha, repo, href }} {...props} />
}

export interface IPullRequestLink {
  number?: number
  repo?: string
  href?: string
  label?: string
  showIcon?: boolean
  className?: string
  textVariant?: TTextVariant
  weight?: TTextWeight
  variant?: 'default' | 'inline'
}

export const PullRequestLink = ({
  number,
  repo,
  href,
  label,
  ...props
}: IPullRequestLink) => {
  if (number == null) return null
  return (
    <GitReferenceLink
      reference={{ type: 'pull-request', number, repo, href, label }}
      {...props}
    />
  )
}

export const TagLink = ({
  tag,
  repo,
  href,
  ...props
}: {
  tag?: string
  repo?: string
  href?: string
  showIcon?: boolean
  className?: string
  textVariant?: TTextVariant
  weight?: TTextWeight
  variant?: 'default' | 'inline'
}) => {
  if (!tag) return null
  return <GitReferenceLink reference={{ type: 'tag', tag, repo, href }} {...props} />
}
