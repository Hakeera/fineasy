# Estágio 1 — compila o Go (binário estático, sem dependências)
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o fineasy ./cmd/cli

# Estágio 2 — imagem final com Go compilado + Python
FROM python:3.12-slim
WORKDIR /app
COPY --from=builder /app/fineasy .
COPY extractor.py .
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
CMD ["./fineasy"]
