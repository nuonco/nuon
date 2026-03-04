import { Loading } from '@/components/common/Loading'

export const ProviderLoading = () => (
  <div className="flex items-center justify-center w-full h-screen">
    <Loading variant="large" className="h-24 w-24 opacity-50 animate-pulse" />
  </div>
)
