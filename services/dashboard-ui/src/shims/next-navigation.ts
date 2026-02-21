import {
  useLocation,
  useNavigate,
  useParams as useRouterParams,
} from 'react-router-dom'

export function usePathname() {
  return useLocation().pathname
}

export function useSearchParams() {
  const { search } = useLocation()
  return new URLSearchParams(search)
}

export function useParams<T extends Record<string, string> = Record<string, string>>(): T {
  return useRouterParams() as T
}

export function useRouter() {
  const navigate = useNavigate()
  return {
    push: (url: string) => navigate(url),
    replace: (url: string) => navigate(url, { replace: true }),
    back: () => navigate(-1),
    refresh: () => window.location.reload(),
    prefetch: () => {},
  }
}

export function redirect(url: string): never {
  window.location.href = url
  throw new Error('redirect')
}

export function notFound(): never {
  throw new Error('NOT_FOUND')
}
