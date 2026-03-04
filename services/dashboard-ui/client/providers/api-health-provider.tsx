import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { getAPIHealth } from '@/lib'
import type { TAPIHealth } from '@/types'

type APIHealthContextValue = {
  health: TAPIHealth
  isLoading: boolean
  error: any
}

export const APIHealthContext = createContext<
  APIHealthContextValue | undefined
>(undefined)

export function APIHealthProvider({
  children,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { isLoading: isUserLoading, isAdmin } = useAuth()
  const {
    data: health,
    error,
    isLoading,
  } = useQuery({
    queryKey: ['api-health'],
    queryFn: getAPIHealth,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  return (
    <APIHealthContext.Provider
      value={{
        health: health!,
        isLoading,
        error,
      }}
    >
      {health?.status === 'degraded' && !isUserLoading ? (
        <Banner className="!rounded-none" theme="error">
          <div className="flex items-center gap-8">
            {isAdmin ? (
              health?.degraded?.length ? (
                health?.degraded?.map((d) => (
                  <DegradedBanner
                    key={d}
                    heading={
                      DEGRADED_MESSAGE[d as keyof typeof DEGRADED_MESSAGE]?.heading ||
                      DEGRADED_MESSAGE['generic']?.heading
                    }
                    message={
                      DEGRADED_MESSAGE[d as keyof typeof DEGRADED_MESSAGE]?.message ||
                      DEGRADED_MESSAGE['generic']?.message
                    }
                  />
                ))
              ) : (
                <GenericDegradedBanner />
              )
            ) : (
              <GenericDegradedBanner />
            )}
          </div>
        </Banner>
      ) : null}
      {children}
    </APIHealthContext.Provider>
  )
}

const DEGRADED_MESSAGE = {
  generic: {
    heading: "We're currently experiencing degraded performance.",
    message:
      'You may notice slower response times or intermittent connectivity issues. Our team is actively working to resolve this issue. We apologize for any inconvenience and appreciate your patience.',
  },
  ch: {
    heading: 'Clickhouse',
    message: 'Unable to access Clickhouse',
  },
  psql: {
    heading: 'Database',
    message: 'Unable to access database',
  },
  temporal: {
    heading: 'Temporal',
    message: 'Unable to access Temporal',
  },
}

const GenericDegradedBanner = () => (
  <DegradedBanner
    heading={DEGRADED_MESSAGE['generic']?.heading}
    message={DEGRADED_MESSAGE['generic']?.message}
  />
)

const DegradedBanner = ({
  heading,
  message,
}: {
  heading: string
  message: string
}) => (
  <div className="flex flex-col">
    <Text variant="base" weight="strong">
      {heading}
    </Text>
    <Text className="max-w-xl" variant="subtext" theme="neutral">
      {message}
    </Text>
  </div>
)
