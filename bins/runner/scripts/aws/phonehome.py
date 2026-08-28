# Phone-home Lambda source. Canonical home is nuonco/runner scripts/aws/phonehome.py
# (fetched by ctl-api via a pinned tag when rendering install stack templates); this
# copy is the staging source for the next tag. Based on aws-v0.1.4, adding the S3
# rendezvous used by customer-managed installs.
#
# Modes, chosen by the NUON_PHONE_HOME_S3_* Lambda environment variables (which the
# template binds to the PhoneHomeS3Bucket/PhoneHomeS3Key stack parameters):
#   - unset (connected installs): POST the payload to the control plane, as before.
#   - set (customer-managed installs): write the payload JSON to the customer's S3 bucket
#     and skip the control-plane POST entirely. The customer-managed runner polls the same
#     object to late-bind its plans to this environment's stack outputs.
import json
import os
import time

import urllib3
import cfnresponse

http = urllib3.PoolManager()

MAX_RETRIES = 5
BASE_DELAY = 1.75


# A stable physical ID keeps CloudFormation from treating every property
# update as a replacement; cfnresponse otherwise defaults it to the log
# stream name, and the replacement's trailing Delete event would clobber
# the S3 rendezvous object right after the Update wrote it.
PHYSICAL_RESOURCE_ID = "nuon-phone-home"


def finish(event, context, error):
    if error is None:
        cfnresponse.send(event, context, cfnresponse.SUCCESS, {}, PHYSICAL_RESOURCE_ID)
    elif event["RequestType"] in ["Create", "Update"]:
        cfnresponse.send(event, context, cfnresponse.FAILED, {"Error": error}, PHYSICAL_RESOURCE_ID)
    else:
        # It's OK if notifying Nuon fails on deletion
        cfnresponse.send(event, context, cfnresponse.SUCCESS, {}, PHYSICAL_RESOURCE_ID)


def post_with_retries(url, encoded_data):
    last_error = None
    for attempt in range(MAX_RETRIES):
        try:
            response = http.request(
                "POST",
                url,
                body=encoded_data,
                headers={"Content-Type": "application/json"},
            )
            if 200 <= response.status < 300:
                print("Response: ", response.data)
                return None
            last_error = f"HTTP {response.status}: {response.data}"
            print(f"Attempt {attempt + 1}/{MAX_RETRIES} failed: {last_error}")
        except Exception as e:
            last_error = str(e)
            print(f"Attempt {attempt + 1}/{MAX_RETRIES} error: {last_error}")

        if attempt < MAX_RETRIES - 1:
            delay = BASE_DELAY * (2**attempt)
            print(f"Retrying in {delay}s...")
            time.sleep(delay)

    print("All retries exhausted. Error: ", last_error)
    return last_error


def lambda_handler(event, context):
    props = event["ResourceProperties"]
    props["request_type"] = event["RequestType"]
    encoded_data = json.dumps(props).encode("utf-8")

    s3_bucket = os.environ.get("NUON_PHONE_HOME_S3_BUCKET", "")
    s3_key = os.environ.get("NUON_PHONE_HOME_S3_KEY", "")
    if s3_bucket and s3_key:
        if event["RequestType"] == "Delete":
            # The rendezvous object describes the live environment; a delete
            # (including the cleanup delete after a resource replacement) must
            # not overwrite outputs the runner still late-binds against.
            print("skipping stack outputs write on delete")
            finish(event, context, None)
            return
        try:
            import boto3

            boto3.client("s3").put_object(
                Bucket=s3_bucket,
                Key=s3_key,
                Body=encoded_data,
                ContentType="application/json",
            )
            print(f"wrote stack outputs to s3://{s3_bucket}/{s3_key}")
            finish(event, context, None)
        except Exception as e:
            print("S3 write failed: ", e)
            finish(event, context, str(e))
        return

    finish(event, context, post_with_retries(props["url"], encoded_data))
