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
	cr package ./helm/kube-copilot
	cr upload --owner feiskyer --git-repo kube-copilot --packages-with-index --token $(GITHUB_TOKEN) --push --skip-existing
	cr index --owner feiskyer --git-repo kube-copilot  --packages-with-index --index-path . --token $(GITHUB_TOKEN) --push

.PHONY: clean
clean:
	rm -rf dist

.PHONY: install-dev
install-dev:
	poetry install

.PHONY: install-poetry
install-poetry:
	curl -sSL https://install.python-poetry.org | python3 -
