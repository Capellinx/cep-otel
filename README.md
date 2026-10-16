# Clima por CEP com Observabilidade (OTEL + Zipkin)

Sistema distribuído em Go com dois microsserviços que cooperam para retornar o clima de uma cidade a partir de um CEP, com rastreamento distribuído via OpenTelemetry e visualização no Zipkin.

- **Serviço A (Input):** recebe o CEP via POST, valida e encaminha para o Serviço B.
- **Serviço B (Orquestração):** consulta a cidade (ViaCEP), a temperatura (WeatherAPI) e retorna Celsius, Fahrenheit e Kelvin.
- **OTEL Collector + Zipkin:** coleta e visualização dos traços.

## Pré-requisito: chave da WeatherAPI

O Serviço B usa a [WeatherAPI](https://www.weatherapi.com/) (chave gratuita). Crie um arquivo `.env` na raiz (baseado em `.env.example`):

```env
WEATHER_API_KEY=sua_chave_aqui
```

## Executar

```bash
docker compose up --build
```

Isso sobe os quatro componentes: Serviço A, Serviço B, OTEL Collector e Zipkin.

## Portas

| Componente | Porta |
|---|---|
| Serviço A (HTTP, entrada) | `8080` |
| Serviço B (HTTP, orquestração) | `8081` |
| OTEL Collector (OTLP gRPC / HTTP) | `4317` / `4318` |
| Zipkin (UI) | `9411` |

## Como fazer a requisição (Serviço A)

A requisição vai para o **Serviço A** na porta `8080`, via `POST`:

```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{"cep":"29902555"}'
```

Resposta de sucesso (`200 OK`):

```json
{
  "city": "São Paulo",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

### Respostas de erro

| Status | Mensagem | Quando |
|---|---|---|
| `422` | `invalid zipcode` | CEP não tem 8 dígitos ou não é string |
| `404` | `can not find zipcode` | CEP bem formado, mas inexistente |

O arquivo [`api.http`](./api.http) contém requisições prontas para todos os casos.

## Observabilidade (Zipkin)

Abra o Zipkin em **http://localhost:9411**, clique em **Run Query** (ou busque por `serviceName=service-a`) e abra um traço.

Cada requisição gera um traço único que atravessa os dois serviços:

```
Request → service-a → service-b
                         ├── busca-cep-localizacao   (span manual: consulta ViaCEP)
                         └── busca-temperatura-clima  (span manual: consulta WeatherAPI)
```

Os spans manuais de busca de CEP e de temperatura medem o tempo de cada chamada externa, e o contexto de trace é propagado de A para B automaticamente (W3C Trace Context), de modo que todo o fluxo aparece em um único traço.

## Arquitetura

```
cmd/servicea/            entrypoint do Serviço A
cmd/serviceb/            entrypoint do Serviço B
internal/servicea/       validação do CEP e encaminhamento para o B
internal/serviceb/       orquestração (localização + clima + conversões)
internal/viacep/         client ViaCEP
internal/weather/        client WeatherAPI
internal/temperature/    conversões de temperatura
internal/observability/  inicialização do OpenTelemetry (exporter OTLP)
otel-collector-config.yaml   pipeline OTLP -> Zipkin
```

## Testes

```bash
go test ./...
```

Cobrem validação do CEP e não-encaminhamento no Serviço A, e os casos 422 / 404 / 200 no Serviço B.

## Observação sobre a fórmula de Kelvin

O enunciado traz a fórmula `K = C + 273`, enquanto o exemplo de resposta usa `273.15` (`28.5°C → 301.65K`). Esta implementação segue a **fórmula explícita (`K = C + 273`)**; para usar `273.15`, basta ajustar `internal/temperature/temperature.go`.
