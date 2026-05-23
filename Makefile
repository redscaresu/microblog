TF_DIR := terraform
TF_VARS_FILE := $(TF_DIR)/terraform.auto.tfvars
TF_VARS_EXAMPLE := $(TF_DIR)/terraform.auto.tfvars.example

.PHONY: tfvars-check tfvars-init tf-init tf-plan tf-apply

tfvars-check:
	@test -f $(TF_VARS_FILE) || { \
		echo "Missing $(TF_VARS_FILE). Run 'make tfvars-init' and edit the file first."; \
		exit 1; \
	}

tfvars-init:
	@test -f $(TF_VARS_FILE) || cp $(TF_VARS_EXAMPLE) $(TF_VARS_FILE)
	@echo "Local Terraform vars ready at $(TF_VARS_FILE)"

tf-init:
	cd $(TF_DIR) && terraform init

tf-plan: tfvars-check
	cd $(TF_DIR) && terraform plan

tf-apply: tfvars-check
	cd $(TF_DIR) && terraform apply
