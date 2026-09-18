import { Card } from '@/components/common/Card'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import {
  BranchRunCommit,
  type IBranchRunCommit,
} from '@/components/branches/BranchRunCommit'

export interface IBranchTrackingCard {
  repo?: string
  branch?: string
  directory?: string
  latestRun?: IBranchRunCommit
}

const repoHref = (repo?: string) => {
  if (!repo) return undefined
  return repo.startsWith('http') ? repo : `https://github.com/${repo}`
}

export const BranchTrackingCard = ({
  repo,
  branch,
  directory,
  latestRun,
}: IBranchTrackingCard) => {
  const directoryPath =
    directory && directory !== '.' && directory !== '/' ? directory : undefined

  if (!repo && !branch && !directoryPath && !latestRun) return null

  const href = repoHref(repo)

  return (
    <Card className="gap-4">
      <Text weight="strong">Tracking</Text>
      <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
        {repo ? (
          <LabeledValue label="Repository">
            {href ? (
              <Link href={href} isExternal>
                {repo}
              </Link>
            ) : (
              <Text variant="subtext">{repo}</Text>
            )}
          </LabeledValue>
        ) : null}
        {branch ? (
          <LabeledValue label="Branch">
            <Text variant="subtext">{branch}</Text>
          </LabeledValue>
        ) : null}
        {directoryPath ? (
          <LabeledValue label="Directory">
            <Text variant="subtext" className="break-all">
              {directoryPath}
            </Text>
          </LabeledValue>
        ) : null}
        {latestRun ? (
          <LabeledValue label="Commit">
            <BranchRunCommit {...latestRun} />
          </LabeledValue>
        ) : null}
      </div>
    </Card>
  )
}
