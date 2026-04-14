# AWS Instance Identity Document (IID) Certificates

These are the AWS public certificates used to verify PKCS7 signatures
on EC2 Instance Identity Documents. Each file is named `<region>.pem`
and contains the RSA 2048 certificate for that region.

## Source

Download from the AWS docs:
https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/regions-certs.html

Use the **RSA-2048** certificates (not RSA, DSA, or other variants).

## Updating

If AWS rotates certificates for a region:

1. Download the new PEM from the link above.
2. Replace the file here and redeploy, **or**
3. Use the config overlay: set `AWS_IID_CERTS_DIR` to a directory
   containing the updated PEM files. The cert store loads from that
   directory first and falls back to these embedded certs for any
   regions not overridden.

## Adding a new region

Add a file named `<region>.pem` to this directory. The region name is
derived from the filename (minus the `.pem` extension).
