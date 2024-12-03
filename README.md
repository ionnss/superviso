# Comandos para iniciar o projeto
## Parar os containers
```shell
docker compose down
```

## Remover o volume do banco de dados para limpar os dados
```shell
docker compose down -v
```

## Reconstruir e iniciar
```shell
docker compose up --build
```
EM CASOS ESPECIAIS, PARA REMOVER TODAS AS IMAGENS RELACIONADAS AO PROJETO:
### Remover todas as imagens relacionadas ao projeto
```shell
docker rmi $(docker images -q superviso_app)
docker rmi $(docker images -q postgres:16.4)
```

### Limpar containers parados, networks não usadas, imagens e volumes
```shell
docker system prune -a --volumes
```

---

# Comando para acessar o container do banco de dados
```shell
docker exec -it superviso-db-1 bash
```
```shell
psql -U osivrepus_ions -d superviso

```
ou somente:
```shell
docker-compose exec db psql -U osivrepus_ions -d superviso
```
```shell
\dt
```
Para sair:
```shell
\q
```

# Supervisão Online - Sistema de Gerenciamento de Supervisão Psicológica

## Sobre o Projeto
O Supervisão Online é uma plataforma web desenvolvida para facilitar a conexão entre supervisores e supervisionados na área de psicologia. O sistema permite que profissionais supervisores disponibilizem horários para supervisão e que psicólogos supervisionados possam agendar sessões de supervisão.

## Funcionalidades Principais

### Para Supervisores
- Cadastro de conta com informações profissionais (CRP, abordagem teórica, qualificações)
- Gerenciamento de disponibilidade de horários
- Definição de valores por sessão
- Visualização e gestão de agendamentos

### Para Supervisionados
- Cadastro de conta com informações profissionais
- Busca de supervisores disponíveis
- Agendamento de sessões de supervisão
- Histórico de supervisões realizadas

## Tecnologias Utilizadas
- Backend: Go (Golang)
- Frontend: HTML, CSS (Bootstrap), JavaScript
- Banco de Dados: PostgreSQL
- Containerização: Docker
- HTMX para interações dinâmicas

## Estrutura do Banco de Dados
O sistema utiliza as seguintes tabelas principais:
- `supervisor`: Armazena dados dos supervisores
- `supervisionated`: Armazena dados dos supervisionados
- `supervisor_availability`: Gerencia disponibilidade de horários

## Segurança
- Sistema de autenticação com proteção contra tentativas de login
- Senhas criptografadas com bcrypt
- Proteção contra bloqueio de conta após múltiplas tentativas falhas

## Requisitos do Sistema
- Docker e Docker Compose instalados
- Conexão com internet para recursos externos (Bootstrap, etc.)

## Configuração de Desenvolvimento
1. Clone o repositório
2. Execute `docker compose up --build` para iniciar o ambiente
3. Acesse `http://localhost:8080` no navegador

## Contribuição
Para contribuir com o projeto:
1. Faça um fork do repositório
2. Crie uma branch para sua feature
3. Faça commit das alterações
4. Envie um pull request
