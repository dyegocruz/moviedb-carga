# Contexto e Diretrizes do Projeto (Go / GoLang)

## Stack Tecnológica
- **Linguagem:** Go (GoLang) 1.24+
- **Banco de Dados Documental:** MongoDB (Driver oficial `go.mongodb.org/mongo-driver`)
- **Busca e Observabilidade:** Elasticsearch (Cliente oficial `github.com/olivere/elastic`)
- **Mensageria / Eventos:** RabbitMQ (Driver `github.com/rabbitmq/amqp091-go`)

## Princípios de Código (Ponytail & Go Idioms)
1. **Padrão Go Idiomático:** Escreva código simples, explícito e legível. Evite abstrações e interfaces desnecessárias.
2. **Reuse & Standard Library:** Dê preferência à biblioteca padrão de Go (`net/http`, `context`, `time`, `encoding/json`, `fmt`) e pacotes já instalados no `go.mod`.
3. **Tratamento de Erros:** Trate os erros explicitamente onde eles ocorrem (`if err != nil`). Retorne erros envelopados com contexto quando útil (`fmt.Errorf("falha ao buscar documento: %w", err)`).
4. **Context Propagation:** Sempre passe e respeite o `ctx context.Context` em operações de I/O (I/O, MongoDB, Elasticsearch, RabbitMQ, chamadas HTTP).
5. **YAGNI (You Aren't Gonna Need It):** Escreva apenas o código estritamente necessário para cumprir o requisito da tarefa.

---

## Arquitetura e Estrutura de Pastas
O projeto usa pacotes na raiz do módulo para as áreas de domínio e serviços, com os executáveis em `cmd/`. A inicialização e a integração entre esses pacotes ficam em `internal/app/`; não há uma divisão atual em `domain/`, `repository/`, `messaging/` ou `handler/`.

- `cmd/catalog/`: executável do processo de carga agendada do catálogo.
- `cmd/catalog-worker/`: executável do consumidor de mensagens de processamento do catálogo.
- `internal/app/`: inicializa os serviços e coordena os fluxos dos executáveis de catálogo e worker.
- `bootstrap/`: carrega a configuração e inicializa os recursos necessários à aplicação, incluindo MongoDB.
- `catalog-charge/`: lógica de carga e atualização do catálogo, incluindo a integração com Elasticsearch.
- `movie/`, `tv/` e `person/`: serviços e tipos relacionados aos catálogos de filmes, séries e pessoas; obtêm dados do TMDB e trabalham com MongoDB.
- `parameter/`: leitura e gerenciamento de parâmetros da aplicação armazenados no MongoDB.
- `tmdb/`: cliente e tipos usados na comunicação com a API do TMDB.
- `services/`: configuração e integrações com MongoDB, Elasticsearch e RabbitMQ.
- `configs/`: leitura e fornecimento da configuração da aplicação, incluindo variáveis de ambiente.
- `common/`: constantes e tipos compartilhados entre pacotes.
- `queue/`: nomes de filas e tipos das mensagens de processamento.
- `util/`: funções utilitárias compartilhadas.
- `vendor/`: dependências Go vendorizadas.

Os testes ficam junto aos pacotes que exercitam, em arquivos `*_test.go`.

---

## Convenções por Tecnologia

### 1. MongoDB (`services/`)
- Utilize structs fortemente tipadas com tags `bson:"nome_campo,omitempty"` e `json:"nome_campo"`.
- Prefira usar a ID padrão `primitive.ObjectID` para a chave `_id`.
- Sempre defina *timeouts* usando `context.WithTimeout` ao executar queries ou mutações.
- Certifique-se de fechar os cursores com `defer cursor.Close(ctx)`.

### 2. Elasticsearch (`services/`)
- Mantenha a conversão entre structs e JSON explícita através de `json.Marshal` / `json.Unmarshal` ou `bytes.Buffer`.
- Trate as respostas checando o código de status HTTP (`res.IsError()`) antes de tentar decodificar a resposta.
- Garanta a leitura e o fechamento do corpo da resposta com `defer res.Body.Close()`.

### 3. RabbitMQ (`services/`)
- Garanta que as conexões e canais sejam gerenciados corretamente com reconexão resiliente, se necessário.
- Mantenha a declaração de Filas (*Queues*) e Trocas (*Exchanges*) idempotente (`QueueDeclare` / `ExchangeDeclare`).
- Consumidores devem processar mensagens de forma assíncrona respeitando acoplamento gracioso e enviando `Ack` ou `Nack`/`Reject` explicitamente baseando-se no sucesso do processamento.
- Use `auto-ack: false` para evitar perda de dados em falhas de consumidor.

---

## Regras de Segurança e Resiliência
- **Segredo / Variáveis de Ambiente:** Nunca insira senhas, URIs de banco ou credenciais hardcoded. Use variáveis de ambiente (ex: lendo via `os.Getenv` ou pacotes como `godotenv`/`viper`).
- **Graceful Shutdown:** Certifique-se de que conexões do MongoDB, Elasticsearch e RabbitMQ fechem adequadamente no encerramento da aplicação (`os.Interrupt`, `syscall.SIGTERM`).
- **Sem Dependências Desnecessárias:** Não adicione novos frameworks (ex: Gin, Fiber, Echo, ORMs complexos) a menos que explicitamente solicitado. Priorize a `net/http` e drivers oficiais.
