#deploy: ## Deploy to Raspbery Pi
#	rsync -rv . 192.168.1.42:apps/go-greenhouse-gpio/ ;\
#	ssh 192.168.1.42 "cd apps/go-greenhouse-gpio && docker-compose build && \
#	docker-compose up -d"
	
deploy: 
	docker build -t noodlensk/go-greenhouse-gpio . && \
	docker push noodlensk/go-greenhouse-gpio && \
	rsync -rv . 192.168.1.42:apps/go-greenhouse-gpio/ && \
	ssh 192.168.1.42 "cd apps/go-greenhouse-gpio && docker-compose pull && docker-compose up -d"
# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help