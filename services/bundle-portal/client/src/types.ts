export type TCatalogRef = {
  id: string
  kind: string
  name: string
  component?: string
  cron_schedule?: string
  steps?: number
}

export type TCatalog = {
  deployment_id: string
  bundle_digest: string
  generated_at: string
  refs?: TCatalogRef[]
}

export type TComponentHealth = {
  component_id?: string
  install_component_id?: string
  component_name?: string
  component_type?: string
  health: string
}

export type THealth = {
  latest?: {
    observed_at?: string
    components?: TComponentHealth[]
  } | null
}

export type TDrift = {
  drifted: boolean
  resource_changes: number
  output_changes: number
  resource_drift: number
  summary?: string
}

export type TRunStep = {
  id: string
  name: string
  kind: string
  job_id?: string
  status: string
  error?: string
  drift?: TDrift
}

export type TRun = {
  run_id: string
  dispatch_id?: string
  ref_id: string
  ref_kind: string
  ref_name: string
  source: string
  status: string
  error?: string
  started_at: string
  finished_at?: string
  steps?: TRunStep[]
}
