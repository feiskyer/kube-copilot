# make run ARGS='execute "find all kube-dns pods"'

.PHONY: run
run:
	poetry run kube-copilot $(ARGS)

.PHONY: build
build:
	poetry build

.PHONY: publish
publish: build
	poetry publish

.PHONY: clean
clean:
	rm -rf dist

.PHONY: install-dev
install-dev:
	poetry install

.PHONY: install-poetry
install-poetry:
	curl -sSL https://install.python-poetry.org | python3 -
