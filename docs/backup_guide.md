# Guia de Backup - Superviso

## Visão Geral
O sistema de backup do Superviso é composto por:
- Backup automático diário do banco de dados
- Retenção de 7 dias de histórico
- Armazenamento local dos backups

## Configuração Local

### 1. Estrutura
```bash
superviso/
├── backups/          # Diretório de backups (ignorado pelo git)
├── scripts/
│   ├── backup.sh     # Script principal de backup
│   └── sync_backups.sh   # Script de sincronização
└── docker-compose.yml    # Inclui serviço de backup
```

### 2. Permissões
```bash
chmod +x scripts/backup.sh
chmod +x scripts/sync_backups.sh
chmod 777 backups
```

### 3. Iniciar Serviços
```bash
docker compose up -d
```

## Configuração em Produção (VPS)

### 1. Primeiro Deploy
```bash
# Criar diretórios necessários
mkdir -p backups
chmod 777 backups
chmod +x scripts/*.sh

# Iniciar serviços
docker compose up -d
```

### 2. Configurar Sincronização Externa
```bash
# Gerar chave SSH (se necessário)
ssh-keygen -t rsa -b 4096

# Copiar chave para servidor de backup
ssh-copy-id usuario@servidor-backup

# Testar sincronização
./scripts/sync_backups.sh
```

### 3. Configurar Cron para Sincronização
```bash
# Editar crontab
crontab -e

# Adicionar linha para sincronização diária às 4am
0 4 * * * /caminho/completo/scripts/sync_backups.sh
```

## Operações Comuns

### Backup Manual
```bash
docker compose exec backup /scripts/backup.sh
```

### Restaurar Backup
```bash
# Listar backups disponíveis
ls -l backups/

# Restaurar backup específico
docker compose exec -T db pg_restore -U osivrepus_ions -d superviso < backups/nome_do_backup.dump
```

### Verificar Logs
```bash
# Logs de backup
cat backups/backup.log

# Logs de sincronização
cat backups/sync.log

# Logs do container
docker compose logs backup
```

## Monitoramento

### Verificações Diárias
1. Confirmar criação de novo backup
2. Verificar tamanho dos arquivos
3. Confirmar sincronização externa

### Verificações Semanais
1. Testar restauração de backup
2. Verificar espaço em disco
3. Revisar logs por erros

## Recuperação de Desastre

### 1. Parar Serviços
```bash
docker compose down
```

### 2. Restaurar Backup Mais Recente
```bash
# Identificar último backup
latest_backup=$(ls -t backups/*.dump | head -1)

# Restaurar
docker compose up -d db
docker compose exec -T db pg_restore -U osivrepus_ions -d superviso < $latest_backup
```

### 3. Reiniciar Serviços
```bash
docker compose up -d
```

## Manutenção

### Limpeza Manual
```bash
# Remover backups mais antigos que 7 dias
find backups/ -name "*.dump" -mtime +7 -delete
```

### Verificar Espaço
```bash
du -sh backups/
```

## Troubleshooting

### Backup Falhou
1. Verificar logs: `docker compose logs backup`
2. Confirmar conexão com banco: `docker compose exec backup pg_isready -h db`
3. Verificar permissões: `ls -la backups/`

### Sincronização Falhou
1. Testar conexão SSH: `ssh usuario@servidor-backup`
2. Verificar espaço em disco: `df -h`
3. Testar manualmente: `./scripts/sync_backups.sh` 