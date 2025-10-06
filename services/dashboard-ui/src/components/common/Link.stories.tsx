import { Link } from './Link'

export const Variants = () => (
  <div className="flex flex-col gap-4">
    <Link href="#">Default link</Link>
    <Link href="#" variant="ghost">
      Ghost link
    </Link>
    <Link href="#" variant="nav">
      Nav link
    </Link>
    <Link href="#" variant="breadcrumb">
      Breadcrumb link
    </Link>
  </div>
)

export const States = () => (
  <div className="flex flex-col gap-4">
    <Link href="#" variant="nav">
      Default
    </Link>
    <Link href="#" variant="nav" isActive>
      Active
    </Link>
  </div>
)
