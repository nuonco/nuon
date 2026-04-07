export default {
  title: 'LogStream/LogSeverity',
}

import { LogSeverity } from './LogSeverity'

export const Trace = () => <LogSeverity severityNumber={4} severityText="Trace" />
export const Debug = () => <LogSeverity severityNumber={8} severityText="Debug" />
export const Info = () => <LogSeverity severityNumber={12} severityText="Info" />
export const Warn = () => <LogSeverity severityNumber={16} severityText="Warn" />
export const Error = () => <LogSeverity severityNumber={20} severityText="Error" />
export const Fatal = () => <LogSeverity severityNumber={24} severityText="Fatal" />
