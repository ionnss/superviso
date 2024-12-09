#!/bin/bash
REMOTE_USER="backupuser"
REMOTE_HOST="backup.seudominio.com"
REMOTE_PATH="/home/backupuser/backups/superviso"
LOCAL_PATH="./backups/"

# Sincroniza backups com servidor remoto
rsync -avz --delete $LOCAL_PATH $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH

# Log da sincronização
echo "Backup sincronizado em $(date)" >> $LOCAL_PATH/sync.log 