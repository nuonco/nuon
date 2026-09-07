import { useEffect, useRef, type ReactNode } from 'react'
import { useLocation, useNavigate } from 'react-router'
import { useSurfaces } from '../../hooks/use-surfaces'
import {
  SurfacesProvider,
  type ISurfaceRegistration,
} from '../../providers/surfaces-provider'
import { Text } from '../atoms/Text'
import { SurfaceHost } from '../organisms/surfaces/SurfaceHost'

type TSurfaces = ReturnType<typeof useSurfaces>

const SurfaceLauncher = ({
  open,
  children,
}: {
  open: (surfaces: TSurfaces) => void
  children?: ReactNode
}) => {
  const surfaces = useSurfaces()
  const opened = useRef(false)

  useEffect(() => {
    if (opened.current) return
    opened.current = true
    open(surfaces)
  }, [open, surfaces])

  return (
    <div className="min-h-screen p-8">
      {children ?? (
        <Text color="secondary">
          The page beneath the active surface remains visible.
        </Text>
      )}
    </div>
  )
}

export const SurfaceStory = ({
  open,
  children,
  registrations,
}: {
  open: (surfaces: TSurfaces) => void
  children?: ReactNode
  registrations?: ISurfaceRegistration[]
}) => (
  <SurfacesProvider>
    <SurfaceHost scope="story" registrations={registrations}>
      <SurfaceLauncher open={open}>{children}</SurfaceLauncher>
    </SurfaceHost>
  </SurfacesProvider>
)

export const SurfacePlayground = ({ children }: { children: ReactNode }) => (
  <SurfacesProvider>
    <SurfaceHost scope="playground">
      <div className="min-h-screen p-8">{children}</div>
    </SurfaceHost>
  </SurfacesProvider>
)

const UrlSurfaceInitializer = ({ search }: { search: string }) => {
  const location = useLocation()
  const navigate = useNavigate()
  const initialized = useRef(false)

  useEffect(() => {
    if (initialized.current) return
    initialized.current = true
    navigate(
      {
        pathname: location.pathname,
        search,
        hash: location.hash,
      },
      { replace: true }
    )
  }, [location.hash, location.pathname, navigate, search])

  return (
    <div className="min-h-screen p-8">
      <Text color="secondary">URL-backed surface fixture</Text>
    </div>
  )
}

export const UrlSurfaceStory = ({
  registrations,
  search,
}: {
  registrations: ISurfaceRegistration[]
  search: string
}) => (
  <SurfacesProvider>
    <SurfaceHost scope="url-story" registrations={registrations}>
      <UrlSurfaceInitializer search={search} />
    </SurfaceHost>
  </SurfacesProvider>
)
