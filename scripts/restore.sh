#!/bin/bash

# Verificar se foi fornecido um arquivo de backup
if [ -z "$1" ]; then
    echo "Uso: $0 <arquivo_backup>"
    echo "Backups disponíveis:"
    ls -lh /backups/*.dump
    exit 1
fi

BACKUP_FILE=$1
POSTGRES_HOST="db"
POSTGRES_DB="${POSTGRES_DB}"
POSTGRES_USER="${POSTGRES_USER}"
export PGPASSWORD="${POSTGRES_PASSWORD}"  # Exporta a senha para pg_restore

# Verificar se arquivo existe
if [ ! -f "$BACKUP_FILE" ]; then
    echo "Erro: Arquivo de backup não encontrado"
    exit 1
fi

echo "Iniciando restauração do backup: $BACKUP_FILE"
echo "ATENÇÃO: Isso irá sobrescrever o banco de dados atual!"
read -p "Continuar? (s/N) " -n 1 -r
echo

if [[ $REPLY =~ ^[Ss]$ ]]; then
    echo "Iniciando restauração em $(date)"
    # Restaurar backup
    pg_restore -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -c -F c $BACKUP_FILE

    if [ $? -eq 0 ]; then
        echo "Restauração concluída com sucesso!"
        echo "----------------------------------------" >> /backups/restore.log
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Restauração realizada" >> /backups/restore.log
        echo "Arquivo: $BACKUP_FILE" >> /backups/restore.log
        echo "Status: Sucesso" >> /backups/restore.log
        echo "----------------------------------------" >> /backups/restore.log
    else
        echo "Erro durante a restauração"
        echo "----------------------------------------" >> /backups/restore.log
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Tentativa de restauração" >> /backups/restore.log
        echo "Arquivo: $BACKUP_FILE" >> /backups/restore.log
        echo "Status: Falha" >> /backups/restore.log
        echo "----------------------------------------" >> /backups/restore.log
        exit 1
    fi
else
    echo "Operação cancelada"
    exit 0
fi 