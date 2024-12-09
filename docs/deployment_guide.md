# Guia de Deploy e Atualizações - Superviso

## Visão Geral
Este guia descreve os procedimentos para atualização do sistema em ambiente de produção.

## 1. Preparação

### 1.1. Backup de Segurança
Primeiro, crie o diretório de backups:
```bash
mkdir -p backups
chmod 777 backups  # Permissão necessária para o container escrever
```

Certifique-se que o script de backup tem permissão de execução:
```bash
chmod +x scripts/backup.sh
```

Sempre faça um backup antes de qualquer atualização:
```bash
# Backup manual
docker compose exec backup /scripts/backup.sh

# Verificar se o backup foi criado
ls -l backups/
```

### 1.2. Verificação do Estado Atual
```bash
# Verificar status dos containers
docker compose ps

# Verificar logs recentes
docker compose logs --tail=100
```

## 2. Tipos de Atualização

### 2.0. Modos de Execução

#### Modo Foreground (Desenvolvimento)
```bash
docker compose up --build
```
- Reconstrói as imagens
- Mantém os logs no terminal
- Logs em tempo real
- Ctrl+C para parar os containers

#### Modo Detached (Produção)
```bash
docker compose up --build -d
```
- Reconstrói as imagens
- Executa em background
- Containers continuam rodando após fechar o terminal
- Ideal para ambiente de produção

#### Visualização de Logs (Modo Detached)
```bash
# Ver todos os logs
docker compose logs

# Ver logs em tempo real
docker compose logs -f

# Ver logs de serviços específicos
docker compose logs app
docker compose logs db
docker compose logs backup
```

### 2.1. Atualização Completa (Com Downtime)
Use quando houver mudanças significativas ou alterações no banco de dados:

```bash
# 1. Parar todos os containers
docker compose down

# 2. Puxar alterações do repositório
git pull origin main

# 3. Reconstruir e iniciar containers
docker compose up --build -d

# 4. Verificar logs
docker compose logs -f
```

### 2.2. Atualização Parcial (Mínimo Downtime)
Use para atualizações menores que afetam apenas a aplicação:

```bash
# 1. Puxar alterações
git pull origin main

# 2. Atualizar apenas o container da aplicação
docker compose up --build -d app

# 3. Verificar logs da aplicação
docker compose logs -f app
```

## 3. Verificações Pós-Atualização

### 3.1. Checklist de Verificação
1. Confirmar que todos os containers estão rodando:
```bash
docker compose ps
```

2. Verificar logs por erros:
```bash
docker compose logs --tail=100
```

3. Testar funcionalidades principais:
   - Login/Registro
   - Perfil de usuário
   - Lista de supervisores
   - Sistema de agendamento

### 3.2. Monitoramento
```bash
# Monitorar uso de recursos
docker stats

# Verificar logs em tempo real
docker compose logs -f
```

## 4. Rollback

### 4.1. Em Caso de Problemas
Se a atualização causar problemas:

1. Reverter para versão anterior:
```bash
# Voltar para commit anterior
git reset --hard HEAD^

# Reconstruir containers
docker compose down
docker compose up --build -d
```

2. Restaurar backup se necessário:
```bash
# Listar backups disponíveis
ls -l backups/

# Restaurar último backup
docker compose exec -T db pg_restore -U osivrepus_ions -d superviso < backups/[ultimo_backup].dump
```

## 5. Boas Práticas

### 5.1. Antes da Atualização
- Avisar usuários sobre manutenção programada
- Fazer backup dos dados
- Verificar espaço em disco disponível
- Testar atualizações em ambiente de staging

### 5.2. Durante a Atualização
- Realizar em horários de baixo uso
- Monitorar logs ativamente
- Manter registro das alterações realizadas

### 5.3. Após a Atualização
- Verificar todas as funcionalidades críticas
- Monitorar performance
- Manter backup anterior por 24h
- Documentar problemas encontrados

## 6. Troubleshooting

### 6.1. Problemas Comuns

1. Container não inicia:
```bash
# Verificar logs detalhados
docker compose logs app
```

2. Erro de conexão com banco:
```bash
# Verificar se banco está respondendo
docker compose exec db pg_isready
```

3. Problemas de permissão:
```bash
# Verificar permissões dos arquivos
ls -la
chmod -R 755 .
```

### 6.2. Logs e Diagnóstico
```bash
# Logs de todos os serviços
docker compose logs

# Logs específicos
docker compose logs app
docker compose logs db
docker compose logs backup

# Logs em tempo real
docker compose logs -f
```

## 7. Manutenção Regular

### 7.1. Tarefas Semanais
- Verificar logs de erro
- Monitorar uso de disco
- Revisar backups
- Verificar execução do cron:
  ```bash
  docker compose exec backup cat /backups/cron.log
  docker compose exec backup ls -l /backups
  ```

### 7.2. Tarefas Mensais
- Limpar logs antigos
- Atualizar dependências
- Testar restauração de backup

## 8. Contatos e Suporte

### 8.1. Em Caso de Emergência
- Email: suporte@superviso.com.br
- Tel: [NÚMERO]
- Discord: [LINK]

---

*Última atualização: [DATA]* 