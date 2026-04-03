REGISTRY ?= localhost:5000
TAG ?= latest

.PHONY: build-all build-event-service build-live-dashboard build-analytics-worker

build-all: build-event-service build-live-dashboard build-analytics-worker

build-event-service:
	@echo "Building event-service..."
	cd apps/event-service && docker build -t $(REGISTRY)/event-service:$(TAG) .

build-live-dashboard:
	@echo "Building live-dashboard..."
	cd apps/live-dashboard && docker build -t $(REGISTRY)/live-dashboard:$(TAG) .

build-analytics-worker:
	@echo "Building analytics-worker..."
	cd apps/analytics-worker && docker build -t $(REGISTRY)/analytics-worker:$(TAG) .

push-all:
	docker push $(REGISTRY)/event-service:$(TAG)
	docker push $(REGISTRY)/live-dashboard:$(TAG)
	docker push $(REGISTRY)/analytics-worker:$(TAG)

deploy-platform:
	kubectl create namespace platform-services --dry-run=client -o yaml | kubectl apply -f -
	helm upgrade --install central-pg ./platform/postgresql/helm -n platform-services
	helm upgrade --install pgbouncer ./platform/pgbouncer/helm -n platform-services

onboard-tenant:
	@test -n "$(TENANT)" || (echo "Usage: make onboard-tenant TENANT=alpha" && exit 1)
	helm upgrade --install tenant-$(TENANT) ./platform/tenant-onboarding/helm \
		-f values/tenant-$(TENANT).yaml

deploy-apps:
	@test -n "$(TENANT)" || (echo "Usage: make deploy-apps TENANT=alpha" && exit 1)
	helm upgrade --install event-service ./apps/event-service/helm \
		-n tenant-$(TENANT) -f values/tenant-$(TENANT).yaml
	helm upgrade --install live-dashboard ./apps/live-dashboard/helm \
		-n tenant-$(TENANT) -f values/tenant-$(TENANT).yaml
	helm upgrade --install analytics-worker ./apps/analytics-worker/helm \
		-n tenant-$(TENANT) -f values/tenant-$(TENANT).yaml

clean:
	helm uninstall event-service live-dashboard analytics-worker -n tenant-alpha 2>/dev/null || true
	helm uninstall tenant-alpha 2>/dev/null || true
	helm uninstall pgbouncer central-pg -n platform-services 2>/dev/null || true
