deploy: ## Deploy to Raspbery Pi
	GOOS=linux GOARCH=arm GOARM=7 go build -o myApp ; \
	scp myApp 192.168.1.42: ; \
	rm myApp

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help