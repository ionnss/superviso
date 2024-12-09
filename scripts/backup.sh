#!/bin/bash
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
POSTGRES_HOST="db"
POSTGRES_DB="superviso"
POSTGRES_USER="osivrepus_ions"

# Criar backup
pg_dump -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -F c -f $BACKUP_DIR/superviso_$TIMESTAMP.dump

# Limpar backups antigos (manter últimos 7 dias)
find $BACKUP_DIR -name "*.dump" -mtime +7 -delete

# Log do backup
echo "Backup criado: superviso_$TIMESTAMP.dump" >> $BACKUP_DIR/backup.log 