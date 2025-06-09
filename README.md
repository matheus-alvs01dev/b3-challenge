## 🚀 Visão Geral

O projeto consiste em 2 partes principais:

1. **Ingestão de Dados**:
    - Lê arquivos CSV/TXT com dados de trades da B3.
    - Processa e insere os dados em lote no banco de dados PostgreSQL usando pq.CopyFrom para alta performance.


2. **API REST de Métricas**:
    - Existe 1 endpoint `/ticker-metrics` que retorna um payload com as métricas de trades por ticker.
   ```json
   {
	    "ticker": "TF583R",
	    "max_range_value": 10,
	    "max_daily_volume": 20000
   }
    ```
    
     - As métricas são calculadas a partir dos dados já persistidos no banco de dados.

⸻

## 🛠 Tecnologias
-	Linguagem: Go 1.24
-	Framework HTTP: Echo
-	DB: PostgreSQL
-	Persistência: SQLC (sqlc v1.29.0)
-	Migrations: Goose ou Makefile
-	Container: Docker Compose (PostgreSQL)
-	Testes & Mocks: Uber Mockgen (GOMOCK)
-	Automação: Makefile para tarefas comuns
-   Golanglintci para qualidade de código

⸻

## 📦 Pré-requisitos
-	Go 1.24 instalado
-	Docker e Docker Compose (para subri o banco de dados)
-	Build Essentials (Linux) ou Xcode Command Line Tools (macOS)

⸻

## ⚙️ Como configurar e rodar o projeto

### 1. Copiar variáveis de ambiente
```bash
cp .env.example .env
```

### 2. Subir o banco de dados PostgreSQL
```bash
docker-compose up -d
```

### 3. Popular o banco de dados com o csv da B3
- as tabelas csv podem ser encontradas na pasta `./b3Data`
- para popular o banco, execute:
```bash
make populate-db
```
### 4. Para iniciar o servidor
```bash
make server
```

## Como Testar
 O servidor estará rodando em `http://localhost:8080` ou você pode configurar a porta no arquivo `.env`.
 para acessar o endpoint de métricas, acesse `http://localhost:8080/ticker-metrics`.
 
 filtros disponíveis:
 - `ticker` (ex: `?ticker=TF583R`)(obrigatório)
 - `trade_date` (ex: `?trade_date=2023-10-01`)(opcional)
 
 Voce pode acessar no insomnia ou postman, ou qualquer outro cliente http.
 Basta usar a collection encontrada na pasta `dev/b3-collection.yaml`.


## Testes Unitários
Para rodar os testes unitários, execute:
```bash
make test
```



