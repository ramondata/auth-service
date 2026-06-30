# Guia de Arquivos Kubernetes - auth-service

## 📁 Estrutura

```
k8s/
├── namespace.yaml                 # Isolamento do projeto
├── configmap.yaml                 # Variáveis não-secretas
├── secret.yaml                    # Dados sensíveis (senhas)
├── postgres-pv.yaml               # PV: disco físico
├── postgres-pvc.yaml              # PVC: pedido de disco
├── postgres-deployment.yaml       # PostgreSQL rodando
├── postgres-service.yaml          # Acesso ao PostgreSQL dentro do cluster
├── auth-service-deployment.yaml   # Auth-service rodando
└── auth-service-service.yaml      # Acesso à auth-service
```

---

## 🎯 Ordem de Criação

Quando for deployar, siga essa ordem:

```bash
# 1. Namespace (contexto isolado)
kubectl apply -f k8s/namespace.yaml

# 2. ConfigMap e Secrets (variáveis)
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

# 3. Storage (disco)
kubectl apply -f k8s/postgres-pv.yaml
kubectl apply -f k8s/postgres-pvc.yaml

# 4. PostgreSQL
kubectl apply -f k8s/postgres-deployment.yaml
kubectl apply -f k8s/postgres-service.yaml

# 5. Auth Service (depende do PostgreSQL)
kubectl apply -f k8s/auth-service-deployment.yaml
kubectl apply -f k8s/auth-service-service.yaml

# OU: Apply tudo de uma vez (K8s respeita dependências)
kubectl apply -f k8s/
```

---

## 📋 Explicação de Cada Arquivo

### 1. namespace.yaml
- **O que é**: Isolamento do seu projeto
- **Analogia**: Como um diretório no seu computador
- **Benefício**: Vários projetos podem rodar sem conflitos

### 2. configmap.yaml
- **O que é**: Variáveis não-secretas
- **Vem do Docker Compose**: 
  ```yaml
  environment:
    PORT: "8001"
    POSTGRES_DB: "auth_db"
  ```
- **Pode mudar sem recriar container**: Sim!

### 3. secret.yaml
- **O que é**: Dados sensíveis em base64
- **Base64 é criptografia?**: NÃO! É só encoding.
- **Como gerar**:
  ```bash
  echo -n "auth_password" | base64
  # Resultado: YXV0aF9wYXNzd29yZA==
  ```

### 4. postgres-pv.yaml
- **O que é**: "Eu tenho um disco de 5GB em /data/postgres"
- **Para Minikube**: Precisa criar a pasta:
  ```bash
  minikube ssh
  mkdir -p /data/postgres
  exit
  ```

### 5. postgres-pvc.yaml
- **O que é**: "Eu preciso de 5GB"
- **O que faz**: Liga ao PV (postgres-pv.yaml)

### 6. postgres-deployment.yaml
- **O que é**: PostgreSQL rodando com as configs
- **Coisas importantes**:
  - `replicas: 1` = Uma única instância (PostgreSQL não scale bem)
  - `livenessProbe` = "Tá vivo?" (substitui healthcheck do Docker Compose)
  - `readinessProbe` = "Pronto pra receber requisições?"
  - `volumeMounts` = Usa a PVC para guardar dados

### 7. postgres-service.yaml
- **O que é**: "Como acessar PostgreSQL dentro do cluster?"
- **Tipo ClusterIP**: Só dentro do cluster (não expõe pra fora)
- **Nome postgres**: Outros Pods usam `postgres:5432` pra conectar

### 8. auth-service-deployment.yaml
- **O que é**: Seu app Go rodando
- **Coisas importantes**:
  - `replicas: 2` = Duas instâncias (tolerância a falhas)
  - `imagePullPolicy: Never` = Usa imagem local (Minikube)
  - `DATABASE_URL` = Conecta em `postgres:5432` (resolve via DNS do Service)
  - `livenessProbe` = Checa `/health` (seu endpoint)

### 9. auth-service-service.yaml
- **O que é**: "Como acessar auth-service?"
- **Tipo NodePort**: Acessível de fora via `<NODE_IP>:30001`
- **Ports**:
  - `port: 8001` = Dentro do cluster
  - `targetPort: 8001` = Porta do Pod
  - `nodePort: 30001` = Porta no seu laptop

---

## 🚀 Antes de Deployar

### 1. Build da imagem Docker
```bash
cd /Users/ramon/fiap/auth-service
docker build -t auth-service:latest .

# Se usar Minikube:
minikube image build -t auth-service:latest .
# OU
docker build -t auth-service:latest . && \
  docker save auth-service:latest | minikube image load -
```

### 2. Criar pasta de dados (se Minikube local)
```bash
minikube ssh
mkdir -p /data/postgres
exit
```

### 3. Init do banco (copiar do seu projeto)
O arquivo `./db/init.sql` precisa estar disponível. Você pode:
- Embutir no PVC via ConfigMap
- Ou rodar migration depois

---

## 🔧 Comandos Úteis Depois

### Ver status
```bash
kubectl get all -n auth-system
kubectl get pods -n auth-system -w  # -w = watch (atualiza em tempo real)
```

### Ver logs
```bash
kubectl logs -f deployment/auth-service -n auth-system
kubectl logs -f deployment/postgres -n auth-system
```

### Acessar seu app
```bash
# IP do Minikube
minikube ip  # ex: 192.168.49.2

# Então acessa em seu navegador:
# http://192.168.49.2:30001/health
```

### Deletar tudo
```bash
kubectl delete namespace auth-system
```

---

## ⚠️ Coisas a Observar

1. **ImagePullPolicy: Never** = Local só (Minikube). Na cloud, use `IfNotPresent` ou `Always`
2. **Secret em base64** = Não é seguro. Em produção, use ExternalSecrets
3. **PV com hostPath** = Local só (Minikube). Na cloud, use StorageClass
4. **nodePort: 30001** = Fixa, manual. Na cloud, use LoadBalancer

---

## 🔄 Conversão do docker-compose.yaml

| Docker Compose | Kubernetes |
|---|---|
| `services:` | `Deployment` |
| `image:` | `containers[].image` |
| `environment:` | `ConfigMap` + `Secret` + `env` |
| `ports:` | `Service` |
| `volumes:` | `PersistentVolume` + `PersistentVolumeClaim` |
| `healthcheck:` | `livenessProbe` + `readinessProbe` |
| `depends_on:` | `readinessProbe` (aguarda que esteja pronto) |

---
