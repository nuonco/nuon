import { api } from '@/lib/api'

export type TCatalogQuery = {
  id: string
  name: string
  description: string
  sql: string
  db_type: string
}

export const getQueryCatalog = () =>
  api<{ queries: TCatalogQuery[] }>({ path: 'query-catalog' })

export const runCatalogQuery = (queryId: string) =>
  api<{ query: TCatalogQuery; results: Record<string, any>[]; count: number }>({
    path: `query-catalog/${queryId}/run`,
    method: 'POST',
  })
