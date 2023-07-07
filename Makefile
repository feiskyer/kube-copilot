# make run ARGS='execute "find all kube-dns pods"'

.PHONY: run
run:
	poetry run kube-copilot $(ARGS)

.PHONY: build
build:
	poetry build

.PHONY: install
install: build
	pip install --force-reinstall --no-deps dist/$(shell ls -t dist | head -n 1)

.PHONY: publish
publish: build
	poetry publish

.PHONY: release-helm
release-helm:
	rm -f ./.cr-release-packages/kube-copilot-*.tgz
	helm package ./helm/kube-copilot -d .cr-release-packages
	git checkout gh-pages
	git pull origin gh-pages
	helm repo index .cr-release-packages
	helm repo index --merge index.yaml .cr-release-packages
	cp .cr-release-packages/*.tgz .
	cp .cr-release-packages/index.yaml .
	git add .
	git commit -am 'Update Helm releases'
	git push origin gh-pages
	git checkout main

.PHONY: release
release: publish release-helm
	gh release create v$(shell poetry version -s) --generate-notes

.PHONY: clean
clean:
	rm -rf dist

.PHONY: install-dev
install-dev:
	poetry install

.PHONY: install-poetry
install-poetry:
	curl -sSL https://install.python-poetry.org | python3 -
