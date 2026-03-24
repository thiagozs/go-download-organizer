# Go Download Organizer

CLI em Go para organizar automaticamente arquivos e pastas de uma pasta de downloads por categoria.

O projeto varre o diretório informado, identifica o tipo dos arquivos por extensão, MIME type e heurísticas simples de nome, e move cada item para uma pasta de categoria correspondente.

## Recursos

- Organiza arquivos e diretórios a partir de um diretório de origem.
- Classifica por extensões simples e multi-extensões, como `.tar.gz`.
- Usa detecção por MIME type como fallback.
- Aplica heurísticas por nome para casos comuns.
- Ignora diretórios já organizados.
- Suporta modo `dry-run` para simular a execução sem mover arquivos.
- Exibe progresso inline em uma única linha no terminal.

## Categorias Suportadas

O organizador atualmente distribui itens nas seguintes categorias:

- `images`
- `videos`
- `audio`
- `documents`
- `archives`
- `packages`
- `code`
- `config`
- `scripts`
- `devops`
- `blockchain`
- `data`
- `design`
- `others`

## Como Funciona

1. Lê todos os itens do diretório de origem.
2. Ignora pastas que já tenham nome de categoria conhecida.
3. Para arquivos, determina a categoria com esta prioridade:
   - multi-extensão;
   - extensão simples;
   - MIME type;
   - heurística por nome.
4. Para diretórios, percorre os arquivos internos e escolhe a categoria dominante.
5. Move o item para a pasta final correspondente dentro do diretório de origem.

## Requisitos

- Go `1.26.0`

## Instalação

Clone o repositório e baixe as dependências:

```bash
go mod download
```

Para gerar o binário:

```bash
go build -o dl-organizer .
```

## Uso

Executando com o binário:

```bash
./dl-organizer organize --source ./Downloads
```

Executando diretamente com Go:

```bash
go run . organize --source ./Downloads
```

Simulação sem mover arquivos:

```bash
go run . organize --source ./Downloads --dry-run
```

## Flags

- `--source`, `-s`: diretório que será organizado. Padrão: `./Downloads`
- `--dry-run`: mostra o que seria feito sem mover os arquivos

## Exemplo de Saída

```text
[==========>.......]  62% (8/13) 📄 contrato.pdf  📚 -> documents
```

## Estrutura do Projeto

```text
.
├── main.go
├── cmd/
│   ├── root.go
│   └── organize.go
└── internal/
    └── organizer/
        ├── organizer.go
        └── classifier.go
```

## Componentes

- `main.go`: ponto de entrada da aplicação.
- `cmd/root.go`: comando raiz da CLI.
- `cmd/organize.go`: comando `organize` e definição das flags.
- `internal/organizer/organizer.go`: execução da varredura, movimentação e progresso no terminal.
- `internal/organizer/classifier.go`: regras de classificação por extensão, MIME e heurística.

## Desenvolvimento

Para validar o projeto:

```bash
go test ./...
```

Formatando o código:

```bash
gofmt -w .
```

## License

Este projeto é distribuido sobre a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## Autor

2026, Thiago Zilli Sarmento :heart:
