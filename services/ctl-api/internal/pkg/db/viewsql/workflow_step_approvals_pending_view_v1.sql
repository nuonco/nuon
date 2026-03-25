SELECT a.*
FROM install_workflow_step_approvals a
LEFT JOIN install_workflow_step_approval_responses resp
  ON resp.install_workflow_step_approval_id = a.id
  AND resp.deleted_at = 0
JOIN install_workflow_steps s
  ON s.id = a.install_workflow_step_id
  AND s.deleted_at = 0
WHERE a.deleted_at = 0
  AND resp.id IS NULL
  AND (s.status->>'status') != 'auto-skipped'
