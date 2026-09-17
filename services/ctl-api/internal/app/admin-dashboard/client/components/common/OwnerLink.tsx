import { Link } from 'react-router'
import { ownerPath } from '@/utils/owner'
import { truncateId } from '@/utils/format'

interface IOwnerLink {
  ownerId?: string
  ownerType?: string
  showType?: boolean
}

export const OwnerLink = ({ ownerId, ownerType, showType = true }: IOwnerLink) => {
  const path = ownerPath(ownerType, ownerId)
  const label = truncateId(ownerId || '')

  return (
    <>
      {path ? (
        <Link to={path} className="font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
          {label}
        </Link>
      ) : (
        <span className="font-mono text-xs">{label}</span>
      )}
      {showType && ownerType && (
        <span className="ml-1 text-[11px] text-gray-400 dark:text-gray-500">({ownerType})</span>
      )}
    </>
  )
}
