#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"

# Use the existing docker-compose-dev.yaml file as a base
cp docker-compose-dev.yaml $1

python3 generate_compose.py $1 $2

