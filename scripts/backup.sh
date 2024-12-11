#!/bin/bash
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
POSTGRES_HOST="db"
POSTGRES_DB="${POSTGRES_DB}"

POSTGRES_USER="${POSTGRES_USER}"
export PGPASSWORD="${POSTGRES_PASSWORD}"  # Exporta a senha para pg_dump

# Verificar se diretório de backup existe
if [ ! -d "$BACKUP_DIR" ]; then
    mkdir -p "$BACKUP_DIR"
    echo "Diretório de backup criado: $BACKUP_DIR"
fi

# Verificar conexão com banco
if ! pg_isready -h $POSTGRES_HOST -U $POSTGRES_USER; then
    echo "Erro: Banco de dados não está acessível"
    exit 1
fi

# Criar backup
echo "Iniciando backup em $(date)"
pg_dump -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -F c -f $BACKUP_DIR/superviso_$TIMESTAMP.dump

# Verificar se backup foi criado com sucesso
if [ $? -eq 0 ]; then
    echo "Backup criado com sucesso: superviso_$TIMESTAMP.dump"
else
    echo "Erro ao criar backup"
    exit 1
fi

# Limpar backups antigos (manter últimos 1 dia)
echo "Removendo backups antigos..."
ls -t $BACKUP_DIR/*.dump | tail -n +4 | xargs rm -f  # Mantém apenas os 3 últimos

# Log do backup
echo "----------------------------------------"
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Iniciando backup" >> $BACKUP_DIR/backup.log
echo "Backup criado: superviso_$TIMESTAMP.dump" >> $BACKUP_DIR/backup.log
echo "Data: $(date)" >> $BACKUP_DIR/backup.log
echo "Tamanho: $(du -h $BACKUP_DIR/superviso_$TIMESTAMP.dump | cut -f1)" >> $BACKUP_DIR/backup.log
echo "Status: $([[ $? -eq 0 ]] && echo "Sucesso" || echo "Falha")" >> $BACKUP_DIR/backup.log
echo "----------------------------------------" >> $BACKUP_DIR/backup.log 