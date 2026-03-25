import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

interface LogJobOutputFilterProps {
  filters: {
    jobOutputOnly: TLogFiltersProps['jobOutputOnly']
    handleJobOutputToggle: TLogFiltersProps['handleJobOutputToggle']
  }
}

export const LogJobOutputFilter = ({ filters }: LogJobOutputFilterProps) => {
  const { jobOutputOnly, handleJobOutputToggle } = filters

  return (
    <CheckboxInput
      labelProps={{
        className: 'h-8',
        labelText: 'Job output',
        labelTextProps: { className: 'font-strong leading-[21px]' },
      }}
      checked={jobOutputOnly}
      onChange={handleJobOutputToggle}
    />
  )
}
