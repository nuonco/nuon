import type { TOTELLog } from '@/types'

export type TLogRecord = Omit<TOTELLog, 'timestamp'> & { timestamp: string }
