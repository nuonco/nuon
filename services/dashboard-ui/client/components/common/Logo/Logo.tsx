import { Link } from '@/components/common/Link'
import { useSidebar } from '@/hooks/use-sidebar'
import { cn } from '@/utils/classnames'
import { LogoDark } from './LogoDark'
import { LogoLight } from './LogoLight'

type TLogoVariant = 'system' | 'dark' | 'light'

interface ILogoBase {
  id?: string
  sidebar?: boolean
  isSidebarOpen?: boolean
  variant?: TLogoVariant
}

const LogoBase = ({
  id,
  sidebar = false,
  isSidebarOpen,
  variant = 'system',
}: ILogoBase) => {
  // Sidebar mode may animate logo or text
  const LOGO_LIGHT_CLASSES = cn('shrink-0', {
    'md:translate-x-[0.55rem]': sidebar && !isSidebarOpen,
    'block dark:hidden': variant === 'system',
    block: variant === 'light',
    hidden: variant === 'dark',
  })
  const LOGO_DARK_CLASSES = cn('shrink-0', {
    'md:translate-x-[0.55rem]': sidebar && !isSidebarOpen,
    'hidden dark:block': variant === 'system',
    hidden: variant === 'light',
    dark: variant === 'dark',
  })
  const LOGO_TEXT_CLASSES = sidebar
    ? cn({
        'md:opacity-100': isSidebarOpen,
        'md:opacity-0': !isSidebarOpen,
      })
    : undefined

  return (
    <Link href="/" className="logo-link w-fit overflow-hidden">
      <span className="sr-only">Nuon</span>
      {variant !== 'dark' ? (
        <LogoLight
          className={LOGO_LIGHT_CLASSES}
          id={id}
          textClassName={LOGO_TEXT_CLASSES}
        />
      ) : null}
      {variant! !== 'light' ? (
        <LogoDark
          className={LOGO_DARK_CLASSES}
          id={id}
          textClassName={LOGO_TEXT_CLASSES}
        />
      ) : null}
    </Link>
  )
}

export const SidebarLogo = () => {
  const { isSidebarOpen } = useSidebar()
  return <LogoBase sidebar isSidebarOpen={isSidebarOpen} />
}

export const Logo = (props: { id?: string; variant?: TLogoVariant }) => (
  <LogoBase {...props} />
)
