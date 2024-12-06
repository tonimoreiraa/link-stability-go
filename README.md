# Ferramenta de monitoramento de rede

Uma ferramenta de monitoramento de rede baseada em Go que realiza testes de latência em vários servidores e gera relatórios JSON detalhados.

## Recursos

- Teste de latência multi-servidor
- Verificações simultâneas de servidor
- Tentativas de repetição configuráveis
- Formato de saída JSON
- Estatísticas detalhadas por endereço

## Instalação

1. Clone o repositório ou crie a seguinte estrutura de diretório:
```
network-monitor/
├── main.go
└── servers.json
```

2. Crie a ferramenta:
```bash
go build -o network-monitor main.go
```

## Configuração

### Configuração do servidor (servers.json)

Crie um arquivo `servers.json` com sua lista de servidores:

```json
[
  {
    "id": 1,
    "name": "Server 1",
    "address": "server1.example.com"
  },
  {
    "id": 2,
    "name": "Server 2",
    "address": "server2.example.com"
  }
]
```

### Constantes

As seguintes constantes podem ser modificadas no código:
- `timeoutDuration`: tempo limite de solicitação HTTP (padrão: 6 segundos)
- `retryCount`: número de tentativas de ping por servidor (padrão: 3)

## Uso

### Uso básico

```bash
./network-monitor 8.8.8.8 1.1.1.1
```

### Arquivo de servidor personalizado

```bash
./network-monitor -servers=custom_servers.json 8.8.8.8
```

### Integração com Zabbix
[Documentação do Zabbix](docs/Zabbix.md)

## Formato de saída

A ferramenta gera resultados no formato JSON:

```json
[
  {
    "address": "1.1.1.1",
    "min_latency_ms": 15.24,
    "max_latency_ms": 45,67,
    "avg_latency_ms": 30,45,
    "timeout_count": 1,
    "online_count": 5,
    "offline_count": 0,
    "total_count": 6,
    "servers": [
      {
        "server_id": 1,
        "server_address": "server1.example.com",
        "responses": [
          {
            "index": 0,
            "type": "online",
            "latency_ms": 15,24,
            "server_id": 1
          }
        ]
      }
    ]
  }
]
```

### Tipos de resposta

- `online`: O servidor respondeu com sucesso
- `timeout`: A resposta do servidor excedeu a duração do tempo limite
- `server-offline`: O servidor de teste não pôde ser alcançado

## Solução de problemas

### Problemas comuns

1. **Erro ao analisar o JSON dos servidores:**
   - Verifique se o formato servers.json corresponde ao exemplo acima
   - Verifique a sintaxe JSON

2. **Nenhum endereço fornecido:**
   - Certifique-se de que pelo menos um endereço IP seja fornecido como argumento

3. **Erro ao ler o arquivo servers:**
   - Verifique se servers.json existe no local correto
   - Verifique as permissões do arquivo