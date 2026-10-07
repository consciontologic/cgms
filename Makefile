# ┊┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃
#  Model-agnostic agent framework Makefile
# ──────────────────────────────────────────────────────────────
#  Targets are thin dispatchers. All real logic lives in
#  xops/makefile/<module>.py (stdlib-only, cross-platform).
#
#  Convention:
#    • daily verbs are short  : help, git
#    • everything else uses   : domain.action  (track.add, git.dry, roadmap.status)
# ┊┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃┃

PYTHON ?= python3
XOPS   := $(PYTHON) xops/makefile

# Tracking append defaults (override on CLI: make track.add ACTION=note SUMMARY="...")
ACTION  ?= note
STATUS  ?= completed
SCOPE   ?= general
AGENT   ?= human
SUMMARY ?=
REFS    ?=
RUN_ID  ?=

.DEFAULT_GOAL := help

.PHONY: help git git.dry track.add track.list roadmap.status codeg

## help              List all available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  make /' | sort

## git               Commit pending tracking rows as conventional commits + push
git:
	@$(XOPS)/git_ops.py push

## git.dry           Preview what `make git` would commit and push (read-only)
git.dry:
	@$(XOPS)/git_ops.py dry

## track.add         Append a row to docs/tracking/tracking.csv (vars: ACTION STATUS SCOPE AGENT SUMMARY REFS RUN_ID)
track.add:
	@$(XOPS)/track_ops.py add \
		--action="$(ACTION)" --status="$(STATUS)" --scope="$(SCOPE)" \
		--agent="$(AGENT)"   --summary="$(SUMMARY)" --refs="$(REFS)" \
		$(if $(RUN_ID),--run-id="$(RUN_ID)",)

## track.list        Show recent tracking rows (last 20)
track.list:
	@$(XOPS)/track_ops.py list

## roadmap.status    Summarize ROADMAP.md checkbox progress
roadmap.status:
	@$(XOPS)/roadmap_ops.py status

## codeg             Initialize or update the CodeGraph index
codeg:
	@$(XOPS)/codegraph_ops.py update

.PHONY: up down restart services.init services.up services.stop services.config service.re go.re web.re
MODE ?= local
SERVICE ?= backend
## up                Build and start web, backend, database and configured related services
up: services.up
## down              Stop the whole stack, retaining containers and persistent volumes
down: services.stop
## restart           Rebuild and recreate the whole configured stack, preserving data
restart: services.up
## services.init     Create private local service secrets (preserves existing)
services.init:
	@$(XOPS)/service_ops.py init
## services.up       Initialize secrets, build Flutter/Go and start configured Compose services
services.up:
	@$(XOPS)/service_ops.py up --mode=$(MODE)
## services.stop     Stop services without removing persistent volumes
services.stop:
	@$(XOPS)/service_ops.py stop --mode=$(MODE)
## services.config   Validate service settings and Compose wiring
services.config:
	@$(XOPS)/service_ops.py config --mode=$(MODE)
## service.re        Recreate SERVICE with bounded readiness and retained logs
service.re:
	@$(XOPS)/service_ops.py restart --service=$(SERVICE) --mode=$(MODE)
## go.re             Build Go and recreate backend, preserving data
go.re:
	@$(XOPS)/service_ops.py go --mode=$(MODE)
## web.re            Build Flutter web and recreate web, preserving data
web.re:
	@$(XOPS)/service_ops.py web --mode=$(MODE)

.PHONY: web.dev
## web.dev           Run Flutter hot reload through the local same-origin proxy
web.dev:
	@$(XOPS)/web_dev.py
