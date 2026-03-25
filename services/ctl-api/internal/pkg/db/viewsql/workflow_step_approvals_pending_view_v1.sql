SELECT a.*
FROM install_workflow_step_approvals a
LEFT JOIN install_workflow_step_approval_responses resp
  ON resp.install_workflow_step_approval_id = a.id
  AND resp.deleted_at = 0
JOIN install_workflow_steps s
  ON s.id = a.install_workflow_step_id
  AND s.deleted_at = 0
JOIN install_workflows w
  ON w.id = s.install_workflow_id
  AND w.deleted_at = 0
  AND w.finished_at IS NULL
  AND (w.status->>'status') NOT IN ('cancelled', 'error')
JOIN installs i
  ON i.id = w.owner_id
  AND i.deleted_at = 0
WHERE a.deleted_at = 0
  AND resp.id IS NULL
  AND (s.status->>'status') NOT IN ('auto-skipped', 'cancelled', 'error')
