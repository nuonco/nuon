# infra-orgs

We manage as much infrastructure related to an org in our `orgs` accounts. Most workers that interact with infrastructure in this account must create cross account IAM roles and "2-step" assume them to lock down permissions as much as possible.

## Resources

This module currently manages:

* installations buckets - buckets for both stage and prod to store installation state and logs
* runs buckets - buckets for managing build and deploy plans and runs

While `installations_bucket` doesn't currently live in the orgs account, it will be migrated shortly.
