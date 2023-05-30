# Builder image
FROM python:3.11-bullseye AS builder

RUN curl -sSL https://install.python-poetry.org | python3 -

WORKDIR /app
COPY . /app

RUN /root/.local/bin/poetry install && /root/.local/bin/poetry build && \
    pip install dist/*.whl


# Final image
FROM python:3.11-bullseye

RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && mv kubectl /usr/local/bin/kubectl && \
    useradd --create-home --shell /bin/bash copilot

COPY --from=builder /app/dist/*.whl /tmp
RUN pip install /tmp/*.whl && rm -f /tmp/*.whl

USER copilot

ENTRYPOINT [ "/usr/local/bin/kube-copilot" ]
