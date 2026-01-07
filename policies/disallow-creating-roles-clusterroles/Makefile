# POLICY_TYPE can be one of `gatekeeper`, `opa`, with different build process
POLICY_TYPE ?= gatekeeper
VERSION ?= $(shell git rev-parse --short HEAD)
SOURCE_FILES := $(shell find . -type f -name '*.rego')

policy.wasm: $(SOURCE_FILES)
ifeq ($(POLICY_TYPE), gatekeeper)
	opa build -t wasm -e policy/violation -o bundle.tar.gz policy.rego
else ifeq ($(POLICY_TYPE), opa)
	opa build -t wasm -e policy/main utility/policy.rego -o bundle.tar.gz policy.rego
else
	@printf "Please assign POLICY_TYPE to either 'gatekeeper' or 'opa'\n"
	exit 1
endif
	tar xvf bundle.tar.gz /policy.wasm
	rm bundle.tar.gz
	touch policy.wasm # opa creates the bundle with unix epoch timestamp, fix it

.PHONY: test
test:
	opa test *.rego


annotated-policy.wasm: policy.wasm metadata.yml
	kwctl annotate -m metadata.yml -u README.md -o annotated-policy.wasm policy.wasm

.PHONY: e2e-tests
e2e-tests: annotated-policy.wasm
	bats e2e.bats

.PHONY: clean
clean:
	rm -f *.wasm *.tar.gz

.PHONY: push
push: annotated-policy.wasm
	aws --no-cli-pager ecr-public get-login-password --region us-east-1 --profile public.NuonAdmin | docker login --username AWS --password-stdin public.ecr.aws
	aws --no-cli-pager ecr-public describe-repositories --repository-names disallow-creating-roles-clusterroles --region us-east-1 --profile public.NuonAdmin > /dev/null 2>&1 || \
	aws --no-cli-pager ecr-public create-repository --repository-name disallow-creating-roles-clusterroles --region us-east-1 --profile public.NuonAdmin
	kwctl push annotated-policy.wasm public.ecr.aws/p7e3r5y0/disallow-creating-roles-clusterroles:$(VERSION)

