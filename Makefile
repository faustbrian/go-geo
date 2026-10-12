GOLIB ?= golib

.PHONY: source-docs-check check check-full ci cohesion config inventory repository-check specification-check specification-check-local workflows

config:
	$(GOLIB) config validate

check: source-docs-check
	$(GOLIB) check --local --all

check-full: source-docs-check config inventory cohesion repository-check specification-check workflows
	$(GOLIB) check --all

ci: source-docs-check config inventory cohesion repository-check specification-check-local workflows check

cohesion:
	$(GOLIB) cohesion check

inventory:
	$(GOLIB) inventory

repository-check:
	$(GOLIB) repository check

specification-check:
	$(GOLIB) specification check --online

specification-check-local:
	$(GOLIB) specification check

workflows:
	$(GOLIB) workflows check

source-docs-check:
	$(GOLIB) source-docs check
