GOLIB ?= golib

.PHONY: check check-full ci cohesion config inventory repository-check specification-check specification-check-local workflows

config:
	$(GOLIB) config validate

check:
	$(GOLIB) check --local --all

check-full: config inventory cohesion repository-check specification-check workflows
	$(GOLIB) check --all

ci: config inventory cohesion repository-check specification-check-local workflows check

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
