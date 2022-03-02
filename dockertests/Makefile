# commands definitions
DOCKER := docker
GO := go

# targets definitions
BTCD_BUILDCONTEXT := ./..
BTCD_DOCKERFILE := image/btcd-dockertests.Dockerfile
BTCD_IMAGENAME := btcd-dockertests

# print setup
GREEN := "\\033[0;32m"
NC := "\\033[0m"
define print
	echo $(GREEN)$1$(NC)
endef

# defaults
default: build

all: image test

# ============
# DEPENDENCIES
# ============

$(DOCKER):
	docker --version

$(GO):
	go version

# ============
# BUILD
# ============

image: $(DOCKER)
	@$(call print, "Building all docker images")
	$(DOCKER) build $(BTCD_BUILDCONTEXT) -f $(BTCD_DOCKERFILE) -t $(BTCD_IMAGENAME)

# =======
# TESTING
# =======

check: test

test: $(DOCKER)
	@$(call print, "Running tests.")
	$(GO) test -test.timeout=20m

# =========
# UTILITIES
# =========

prune: $(DOCKER)
	@$(call print, "Prune containers.$(NC)")
	$(DOCKER) container prune -f
	@$(call print, "Prune images.$(NC)")
	$(DOCKER) image rm $(BTCD_IMAGENAME)

.PHONY: all \
	default \
	image \
	check \
	test \
	prune