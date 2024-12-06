# Integração com Zabbix

## Preparando o Script para o Zabbix

1. **Build do Script**  
   Compile o script Go e salve-o com o nome `net.link.test`:
   ```bash
   go build -o net.link.test main.go
   ```

2. **Mover para o Diretório de Scripts Externos do Zabbix**
   Coloque o binário e arquivo de servidores na pasta de scripts externos do Zabbix e ajuste as permissões de acesso para o Zabbix:
   ```bash
   sudo cp servers.json /usr/lib/zabbix/externalscripts/
   sudo mv net.link.test /usr/lib/zabbix/externalscripts/
   sudo chown zabbix:zabbix /usr/lib/zabbix/externalscripts
   sudo chmod -R 755 zabbix:zabbix /usr/lib/zabbix/externalscripts
   ```

## Configurando o Item no Zabbix

1. **Criar o Item Principal**  
   No frontend do Zabbix, crie um item na configuração do host que deseja monitorar:
   - **Nome do Item**: `net.link.test[{$IP}]`
   - **Tipo**: **Script Externo**
   - **Parâmetros**: Passe o IP do servidor como macro:
     ```
     {$IP}
     ```
   - **Tipo de Informação**: **Texto**

2. **Itens Dependentes**  
   Configure itens dependentes para processar os resultados detalhados do JSON retornado pelo script.

   **Itens Dependentes:**
   - **Offline Count**  
     - **Nome**: `net.link.test.offlinecount`  
     - **Pré-processamento**: JSON Path  
     - **Tipo de Informação**: Númerico (inteiro sem sinal)  
       ```json
       $.offline_count
       ```

   - **Total Requests**  
     - **Nome**: `net.link.test.total`  
     - **Pré-processamento**: JSON Path  
     - **Tipo de Informação**: Númerico (inteiro sem sinal)  
       ```json
       $.total_count
       ```

   - **Online Count**  
     - **Nome**: `net.link.test.onlinecount`  
     - **Pré-processamento**: JSON Path  
     - **Tipo de Informação**: Númerico (inteiro sem sinal) 
       ```json
       $.online_count
       ```

   - **Timeout Count**  
     - **Nome**: `net.link.test.timeoutcount`  
     - **Pré-processamento**: JSON Path
     - **Tipo de Informação**: Númerico (inteiro sem sinal)  
       ```json
       $.timeout_count
       ```

   - **Average Latency**  
     - **Nome**: `net.link.test.average` 
     - **Pré-processamento**: JSON Path
     - **Tipo de Informação**: Númerico (inteiro sem sinal)  
       ```json
       $.avg_latency_ms
       ```

   - **Address**  
     - **Nome**: `net.link.test.address` 
     - **Pré-processamento**: JSON Path
     - **Tipo de Informação**: Texto
       ```json
       $.address
       ```

   - **Min Latency**  
     - **Nome**: `net.link.test.minlatency` 
     - **Pré-processamento**: JSON Path
     - **Tipo de Informação**: Númerico (inteiro sem sinal)  
       ```json
       $.min_latency_ms
       ```

   - **Max Latency**
     - **Nome**: `net.link.test.maxlatency` 
     - **Pré-processamento**: JSON Path
     - **Tipo de Informação**: Númerico (inteiro sem sinal)  
       ```json
       $.max_latency_ms
       ```

---

## Testando a Configuração

1. **Testar o Script Manualmente**  
   Execute o script manualmente para garantir que ele funcione e retorne JSON:
   ```bash
   /usr/lib/zabbix/externalscripts/net.link.test 8.8.8.8

2. **Verificando o Zabbix**
   Após configurar o item e seus dependentes, verifique se os dados estão sendo coletados corretamente.