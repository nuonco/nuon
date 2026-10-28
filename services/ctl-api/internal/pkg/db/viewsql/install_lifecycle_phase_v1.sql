(
    CASE
        WHEN COALESCE({{lifecycle_phase}}->>'phase', '') IN ('', 'unset') THEN 'unset'
        WHEN {{lifecycle_phase}}->>'phase' IN ('provisioning', 'provisioned', 'deprovisioning', 'deprovisioned', 'reprovisioning')
            THEN {{lifecycle_phase}}->>'phase'
        ELSE 'other'
    END
)
