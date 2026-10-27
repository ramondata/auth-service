# auth-service (Go)

Este é o serviço de autenticação do projeto ToggleMaster. Ele é responsável por criar e validar chaves de API.

## 📦 Pré-requisitos (Local)

* [Go](https://go.dev/doc/install) (versão 1.21 ou superior)
* [PostgreSQL](https://www.postgresql.org/download/) (rodando localmente ou em um contêiner Docker)

## 🚀 Rodando Localmente

1.  **Clone o repositório** e entre na pasta `auth-service`.

2.  **Prepare o Banco de Dados:**
    * Crie um banco de dados no seu PostgreSQL (ex: `auth_db`).
    * Execute o script `db/init.sql` para criar a tabela `api_keys`:
        ```bash
        psql -U seu_usuario -d auth_db -f db/init.sql
        ```

3.  **Configure as Variáveis de Ambiente:**
    Crie um arquivo chamado `.env` na raiz desta pasta (`auth-service/`) com o seguinte conteúdo:
    ```.env
    # String de conexão do seu banco de dados PostgreSQL
    DATABASE_URL="postgres://SEU_USUARIO:SUA_SENHA@localhost:5432/auth_db"

    # Porta que o serviço irá rodar
    PORT="8001"

    # Chave mestra para criar novas chaves de API
    MASTER_KEY="admin-secreto-123"
    ```

4.  **Instale as Dependências:**
    ```bash
    go mod tidy
    ```

5.  **Inicie o Serviço:**
    ```bash
    go run .
    ```
    O servidor estará rodando em `http://localhost:8001`.

## 🧪 Testando os Endpoints

Você pode usar `curl` ou Postman.

**1. Verifique a Saúde (Health Check):**
```bash
curl http://localhost:8001/health
```

Saída esperada: `{"status":"ok"}`

**2. Crie uma nova Chave de API (requer a MASTER_KEY):**

```bash
curl -X POST http://localhost:8001/admin/keys \
-H "Content-Type: application/json" \
-H "Authorization: Bearer admin-secreto-123" \
-d '{"name": "meu-primeiro-servico"}'
```

Saída esperada (A SUA CHAVE SERÁ DIFERENTE):

```json
{
  "name": "meu-primeiro-servico",
  "key": "tm_key_a1b2c3d4...",
  "message": "Guarde esta chave com segurança! Você não poderá vê-la novamente."
}
```

**3. Valide a Chave que você acabou de criar: (Substitua tm_key_... pela chave exata que você recebeu no passo anterior)**

```bash
curl http://localhost:8001/validate \
-H "Authorization: Bearer tm_key_a1b2c3d4..."
```

Saída esperada: `{"message":"Chave válida"}`

**4. Teste uma Chave Inválida:**

```bash
curl http://localhost:8001/validate \
-H "Authorization: Bearer chave-errada-123"
```

Saída esperada: `Chave de API inválida ou inativa`

## CI/CD no AWS Academy

A branch `aws-cloud-k8s` autentica com as credenciais temporárias do laboratório.
O workflow recebe um único secret, `AWS_ACADEMY_CREDENTIALS`, sem assumir outra
role via OIDC. As permissões AWS são as da sessão fornecida pelo Academy.

### Configuração inicial

Em **Settings → Secrets and variables → Actions → Secrets**, crie o repository
secret `AWS_ACADEMY_CREDENTIALS` e cole o bloco de **AWS Details → AWS CLI**:

```ini
[default]
aws_access_key_id=COLE_A_ACCESS_KEY_ATUAL
aws_secret_access_key=COLE_A_SECRET_KEY_ATUAL
aws_session_token=COLE_O_SESSION_TOKEN_ATUAL
```

O cabeçalho `[default]` é opcional. Os três valores são obrigatórios.
Mantenha as credenciais somente no secret, fora de arquivos versionados e de
inputs de execução do workflow.

Em **Variables**, configure:

| Variável | Valor |
| --- | --- |
| `AWS_REGION` | Região do ECR no laboratório, por exemplo `us-east-1` |
| `ECR_REPOSITORY` | Nome do repositório, por exemplo `toggle-master/auth-service` |
| `APP_VERSION` | Versão da aplicação, por exemplo `v1.0.0` |

O ECR deve existir nessa região e conta. O secret existente `GITOPS_DEPLOY_KEY`
continua necessário para atualizar `ramondata/toggle-master-gitops`.
`AWS_ROLE_ARN` não é usado pelo auth-service no modo Academy.

### Atualização diária

1. Inicie o laboratório e copie as novas credenciais em **AWS Details → AWS CLI**.
2. Edite somente o valor de `AWS_ACADEMY_CREDENTIALS`, substituindo o bloco inteiro.
3. Em **Actions → Auth Service CI**, abra a execução mais recente iniciada por
   **push** na `aws-cloud-k8s` que já contenha esta alteração e selecione
   **Re-run all jobs**. Reexecutar uma versão antiga mantém o workflow antigo.

Publicações por push na `aws-cloud-k8s` também usam o secret atualizado.
O fluxo valida a sessão e o ECR, constrói e verifica a imagem e publica no ECR.
O GitOps recebe tanto o repositório da conta autenticada quanto a nova tag.
Pull requests executam as verificações de código sem publicar imagens.

O workflow também aceita `workflow_dispatch` na `aws-cloud-k8s`. O botão
**Run workflow** só aparece quando o arquivo existe na branch padrão (`main`);
até lá, use a reexecução acima. Alterar o secret não inicia uma execução.

Se ocorrer `ExpiredToken` ou `InvalidClientTokenId`, copie um novo bloco de
credenciais do laboratório ativo. Se ocorrer `RepositoryNotFoundException`,
confira a região e a existência do ECR na conta do Academy.
