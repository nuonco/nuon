import { DestructiveChangeBanner } from './DestructiveChangeBanner'

export default {
  title: 'Features / Diffs / Destructive changes',
}

const database = {
  id: 'db',
  title: 'aws_db_instance.primary',
  operation: 'replace' as const,
}

const bucket = {
  id: 'bucket',
  title: 'aws_s3_bucket.logs',
  operation: 'delete' as const,
}

export const DeleteAndReplace = () => (
  <DestructiveChangeBanner changes={[database, bucket]} />
)

export const ReplaceOnly = () => (
  <DestructiveChangeBanner changes={[database]} />
)

export const DeleteOnly = () => <DestructiveChangeBanner changes={[bucket]} />

export const None = () => <DestructiveChangeBanner changes={[]} />
