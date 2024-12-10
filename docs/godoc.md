# Documentação GoDoc - Superviso

## Como Visualizar
1. Instalar godoc:
```bash
go install golang.org/x/tools/cmd/godoc@latest
```

2. Iniciar servidor de documentação:
```bash
godoc -http=:6060
```

3. Acessar: `http://localhost:6060/pkg/superviso/`

## Pacotes Documentados

### main
- Ponto de entrada da aplicação
- Configuração do servidor
- Inicialização do banco de dados

### api/auth
- Autenticação JWT
- Middleware de proteção de rotas
- Gerenciamento de contexto

### api/user
- Registro e login de usuários
- Gerenciamento de perfis
- Atualização de dados

### api/supervisor
- Listagem de supervisores
- Filtros e busca
- Gerenciamento de disponibilidade

### db
- Conexão com PostgreSQL
- Execução de migrações
- Gerenciamento de transações

### models
- Estruturas de dados
- Definições de tipos
- Validações

## Convenções de Documentação
1. Comentários de pacote antes da declaração
2. Comentários de função com parâmetros e retornos
3. Exemplos em arquivos _test.go 