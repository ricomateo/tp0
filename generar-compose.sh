#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"

BASE_FILE=temp-compose.yaml

# Restore any changes made to docker-compose-dev.yaml
git restore docker-compose-dev.yaml

# Use the existing docker-compose-dev.yaml file as a base
cp docker-compose-dev.yaml $BASE_FILE

python3 generate_compose.py $BASE_FILE $1 $2

rm $BASE_FILE
