# fineasy

Protótipo de pipeline para leitura automática de e-mails do Gmail, extração de anexos PDF e persistência em banco de dados PostgreSQL.

---

## Visão geral

O `fineasy` autentica via OAuth2 na API do Gmail, busca e-mails com anexos PDF, salva os metadados no banco e faz o download dos arquivos localmente. O objetivo final é processar contas recebidas por e-mail de forma automatizada.

```
Gmail API
    │
    ▼
cmd/cli/main.go          ← ponto de entrada
    ├── internal/auth    ← fluxo OAuth2
    ├── internal/gmail   ← listagem, leitura e extração de anexos
    └── internal/storage ← persistência no PostgreSQL
```

---

## Estrutura de pastas

```
fineasy/
├── cmd/
│   └── cli/
│       └── main.go               # Ponto de entrada da aplicação
├── internal/
│   ├── auth/
│   │   └── ...                   # Fluxo OAuth2 com o Google
│   ├── gmail/
│   │   ├── service.go            # Criação do cliente Gmail
│   │   └── attachments.go        # Extração e download de PDFs
│   └── storage/
│       └── email_repository.go   # Operações no banco de dados
├── migrations/
│   └── 001_init.sql              # Schema inicial
├── data/
│   └── pdfs/                     # PDFs baixados (gerado em runtime)
├── credentials.json              # Credenciais OAuth2 (não versionar)
└── token.json                    # Token de acesso (gerado em runtime, não versionar)
```

---

## Pré-requisitos

- Go 1.21+
- PostgreSQL 14+
- Projeto no Google Cloud com a Gmail API habilitada
- Arquivo `credentials.json` obtido no Console do Google Cloud

---

## Configuração

### 1. Banco de dados

Execute a migration para criar as tabelas:

```bash
psql -U <usuario> -d <banco> -f migrations/001_init.sql
```

Configure a string de conexão via variável de ambiente (ou no código de `storage.NewConnection`):

```bash
export DATABASE_URL="postgres://usuario:senha@localhost:5432/fineasy"
```

### 2. Credenciais do Gmail

Coloque o arquivo `credentials.json` na raiz do projeto. Na primeira execução, o fluxo OAuth2 abrirá o navegador para autorização e salvará o token em `token.json`.

**Atenção:** nunca versione `credentials.json` nem `token.json`. Adicione ambos ao `.gitignore`.

```
credentials.json
token.json
data/
```

### 3. Executar

```bash
go run cmd/cli/main.go
```

---

## Como funciona

1. Lê `credentials.json` e inicializa o cliente OAuth2.
2. Conecta ao PostgreSQL.
3. Busca os últimos 15 e-mails com anexo PDF na caixa de entrada (`in:inbox has:attachment filename:pdf`).
4. Para cada e-mail:
   - Se o `gmail_id` já existe no banco → pula.
   - Se é novo → salva os metadados e extrai os PDFs para `data/pdfs/`.

Os arquivos são salvos com o padrão `data/pdfs/<gmail_id>_<nome_do_arquivo>.pdf`.

---

## Schema do banco

### `emails`

| Coluna | Tipo | Descrição |
|---|---|---|
| `id` | SERIAL | Chave primária |
| `gmail_id` | TEXT UNIQUE | ID único da mensagem no Gmail |
| `subject` | TEXT | Assunto do e-mail |
| `from_email` | TEXT | Remetente |
| `received_at` | TIMESTAMP | Data de recebimento |
| `created_at` | TIMESTAMP | Data de inserção no banco |

### `attachments`

| Coluna | Tipo | Descrição |
|---|---|---|
| `id` | SERIAL | Chave primária |
| `email_id` | INTEGER | FK para `emails.id` |
| `filename` | TEXT | Nome do arquivo |
| `mime_type` | TEXT | Tipo MIME |
| `file_path` | TEXT | Caminho local do arquivo |
| `created_at` | TIMESTAMP | Data de inserção |

> **Nota:** a tabela `attachments` está definida no schema mas ainda não é populada pelo código atual. Está planejada para uma próxima fase do desenvolvimento.

---

## Pendências e bugs conhecidos

| # | Descrição | Arquivo | Prioridade |
|---|---|---|---|
| 3 | `received_at` é inserido como string bruta em vez de `time.Time` | `internal/storage/email_repository.go` | Alta |
| 4 | Tabela `attachments` nunca é populada | `cmd/cli/main.go` | Média |

---

# TODO
- [ ] Registrar anexos na tabela `attachments` após o download
- [ ] Adicionar parsing de `received_at` para `time.Time` (formato RFC1123Z)
- [ ] Processar o conteúdo dos PDFs (OCR / extração de texto)
- [ ] Classificar os PDFs por tipo (extrato bancário, nota fiscal, etc.)
- [ ] Adicionar testes unitários para `gmail` e `storage`
- [ ] Configurar variáveis de ambiente com `godotenv` ou similar
- [ ] Containerizar com Docker + docker-compose (app + PostgreSQL)

---

## Dependências principais

| Pacote | Uso |
|---|---|
| `google.golang.org/api/gmail/v1` | Cliente oficial da Gmail API |
| `golang.org/x/oauth2/google` | Fluxo OAuth2 |
| `github.com/jackc/pgx/v5` | Driver PostgreSQL |


# Python - pdfplumber

## Criação do ambiente

```
# Na raiz do projeto
python -m venv .venv

# Ativar (Windows)
.venv\Scripts\activate

# Ativar (Linux/Mac)
source .venv/bin/activate

# Instalar dependências
pip install pdfplumber

# Salvar o que foi instalado (fundamental para replicar)
pip freeze > requirements.txt
```
## Replicar amibente

```
python -m venv .venv
.venv\Scripts\activate
pip install -r requirements.txt
```
