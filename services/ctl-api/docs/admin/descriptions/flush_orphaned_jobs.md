Occasionally, things such as a non-running event-loop or another type of bug in our system could cause issues where a
job could be stuck in a queud state. This causes a lot of confusion and friction to our users.

This endpoint manually flushes those jobs, and while this will be in the future, automated as part of the job loop, for 
now it happens rarely enough we can just flush them manually.
