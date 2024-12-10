# Documentação GoDoc - Superviso

## Como Visualizar
1. Se não estiver instalado, instalar godoc:
```bash
go install golang.org/x/tools/cmd/godoc@latest
```

2. Iniciar servidor de documentação:
```bash
godoc -http=:6060
```

3. Acessar: `http://localhost:6060/pkg/superviso/`

## Status da Documentação

### ✅ Pacotes Documentados
- [x] main
  - Ponto de entrada da aplicação
  - Configuração do servidor
  - Inicialização do banco de dados

- [x] api/auth
  - jwt.go: Autenticação e middleware

### ⏳ Pacotes Pendentes

#### api/docs
- [ ] handler.go
  - Gerenciamento de documentação legal
  - Formatação de documentos
  - Rotas de documentação

#### api/routes
- [ ] routes.go
  - Configuração de rotas
  - Middleware de autenticação
  - Handlers HTTP

#### api/supervisor
- [ ] supervisor.go
  - Listagem de supervisores
  - Filtros de busca
  - Formatação de dados

#### api/user
- [ ] profile.go
  - Gerenciamento de perfil
  - Atualização de dados
  - Toggle supervisor
- [ ] user.go
  - Registro de usuários
  - Autenticação
  - Gerenciamento de sessão

#### db
- [ ] connection.go
  - Conexão com PostgreSQL
  - Execução de migrações
  - Tratamento de erros

#### models
- [ ] supervisor.go
  - Estrutura de dados do supervisor
  - Tags JSON
  - Validações
- [ ] user.go
  - Estrutura de dados do usuário
  - Tags JSON
  - Validações

## Convenções de Documentação
1. Comentários de pacote antes da declaração `package`
2. Comentários de função com parâmetros e retornos
3. Exemplos em arquivos _test.go

## Status Atual
- ✅ Documentados: 2 pacotes (main, api/auth)
- ⏳ Pendentes: 6 pacotes (api/docs, api/routes, api/supervisor, api/user, db, models)