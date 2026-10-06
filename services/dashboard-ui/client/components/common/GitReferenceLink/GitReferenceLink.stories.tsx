export default {
  title: 'UI / Git reference link',
}

import { CommitLink, PullRequestLink, TagLink } from './GitReferenceLink'

const sha = 'e5aef07c91b24d0a8f3310c0e5aef07c91b24d0a'

export const Commits = () => (
  <div className="flex flex-col gap-3 p-4">
    <CommitLink sha={sha} repo="acme/platform" />
    <CommitLink sha={sha} repo="https://github.com/acme/platform.git" />
    <CommitLink sha={sha} repo="https://gitlab.com/acme/platform" />
    <CommitLink sha={sha} />
  </div>
)

export const PullRequests = () => (
  <div className="flex flex-col gap-3 p-4">
    <PullRequestLink number={104} repo="acme/platform" />
    <PullRequestLink number={104} href="https://github.com/acme/platform/pull/104" />
    <PullRequestLink number={104} />
  </div>
)

export const Tags = () => (
  <div className="flex flex-col gap-3 p-4">
    <TagLink tag="v1.4.2" repo="acme/platform" />
    <TagLink tag="v1.4.2" />
  </div>
)
