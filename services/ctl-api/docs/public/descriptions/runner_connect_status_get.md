# get runner connect status

The connected status is based on runner heartbeat:

if no heart beat found — false
if heart beat > 15 seconds ago — false, hb timestamp
if the heart beat < 15 seconds ago — true