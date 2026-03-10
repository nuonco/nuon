# Restart

Restarts the `runner install` process (`nuon-runner.service`) via systemctl.

This job is sent to the `runner mng` process when the install runner needs to be restarted — for example, after a token
has been invalidated.

Unlike the `shutdown` job (which shuts down `runner mng` itself) or the `update` job (which also writes a new image
config before restarting), this job is solely concerned with restarting the `nuon-runner.service` systemd unit.
