import { useSearchParamState } from '@/hooks/use-search-param-state'

export const useStepTab = () =>
  useSearchParamState('step_tab', { clearOnUnmount: true })
