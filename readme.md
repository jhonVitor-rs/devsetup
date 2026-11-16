# devsetup

Reinstale o seu ambiente de desenvolvimento depois de formatar o computador, a partir de uma única interface no terminal.

O `devsetup` detecta o seu sistema operacional, mostra quais linguagens e ferramentas já estão instaladas e instala as que faltam com um `enter`. Ele é um binário único, sem dependências, então numa máquina recém-formatada basta baixar e executar.

> **Status:** em desenvolvimento. Os manifestos de Manjaro/Arch são os mais testados; os de Ubuntu/Debian, macOS e Windows ainda precisam de validação.

## Sistemas suportados

| Sistema                       | Gerenciador usado |
| ----------------------------- | ----------------- |
| Arch / Manjaro                | `pacman`          |
| Debian / Ubuntu (e derivadas) | `apt`             |
| macOS                         | `brew`            |
| Windows                       | `winget`          |

A distribuição Linux é detectada em tempo de execução pelo `/etc/os-release` (campos `ID` e `ID_LIKE`), então um único binário Linux serve para todas.

## Uso

```bash
devsetup            # abre a interface interativa
devsetup list       # mostra o status das libs sem abrir a interface
```

Atalhos da interface:

| Tecla     | Ação                                       |
| --------- | ------------------------------------------ |
| `↑` / `↓` | navegar                                    |
| `/`       | filtrar a lista                            |
| `enter`   | instalar a lib selecionada                 |
| `c`       | abrir as configurações (sistema detectado) |
| `h`       | alternar a ajuda curta/completa            |
| `esc`     | cancelar / sair                            |

Pontos de comportamento:

- **Senha do `sudo`.** No Linux, quando a lib exige `sudo`, o app pede a senha num campo próprio, valida e a guarda **apenas em memória** durante a sessão. Ela é entregue ao `sudo` pela entrada padrão e nunca é gravada em disco nem passada como argumento.
- **Dependências.** Uma lib pode exigir outras (`require`). Se alguma não estiver instalada, o app avisa na barra de status e não executa nada.
- **Falhas.** O erro e as últimas linhas da saída do comando ficam na barra de status. Se o comando terminar com sucesso mas o `check` não encontrar a lib (comum quando o `PATH` só atualiza num terminal novo), o app avisa isso também.

## Instalação

### Binário pronto

Baixe o arquivo da sua plataforma na página de releases do repositório:

| Plataforma            | Arquivo                      |
| --------------------- | ---------------------------- |
| Linux (x86_64)        | `devsetup-linux-amd64`       |
| Linux (ARM64)         | `devsetup-linux-arm64`       |
| macOS (Intel)         | `devsetup-darwin-amd64`      |
| macOS (Apple Silicon) | `devsetup-darwin-arm64`      |
| Windows (x86_64)      | `devsetup-windows-amd64.exe` |

No Linux e no macOS, dê permissão de execução:

```bash
chmod +x devsetup-linux-amd64
./devsetup-linux-amd64
```

Se o macOS bloquear o arquivo baixado pelo navegador:

```bash
xattr -d com.apple.quarantine devsetup-darwin-arm64
```

No Windows, o SmartScreen pode avisar que o executável não é assinado: clique em "Mais informações" e depois em "Executar assim mesmo". Use de preferência o Windows Terminal.

### A partir do código-fonte

Requer Go 1.22 ou superior.

```bash
git clone https://github.com/jhonVitor-rs/devsetup.git
cd devsetup
go run .              # compila e abre a interface
go build -o devsetup .
```

## Como funciona

Cada lib é descrita por um arquivo TOML em `app/libs/list/`. Esses arquivos são embutidos no binário com `go:embed` durante a compilação (por isso, ao adicionar ou editar um manifesto, é preciso compilar de novo).

Na inicialização, o app:

1. detecta o sistema operacional;
2. carrega e valida os manifestos;
3. roda o `check` de cada lib em paralelo (timeout de 5 segundos cada);
4. abre a interface com o resultado.

Depois de cada instalação, o `check` roda de novo para atualizar o status.

## Adicionando uma lib

Crie um arquivo em `app/libs/list/`:

```toml
# app/libs/list/jq.toml
id = "jq"
name = "jq"
check = "jq --version"
require = []              # opcional: IDs de outras libs

[install]
arc = "sudo pacman -S --needed --noconfirm jq"
deb = "sudo apt install -y jq"
mac = "brew install jq"
windows = "winget install -e --id jqlang.jq --accept-package-agreements --accept-source-agreements"
```

| Campo     | Descrição                                                                                                            |
| --------- | -------------------------------------------------------------------------------------------------------------------- |
| `id`      | identificador único (obrigatório)                                                                                    |
| `name`    | nome exibido na lista (obrigatório)                                                                                  |
| `check`   | comando que indica se a lib está instalada: código de saída `0` significa instalada, e a saída vira a versão exibida |
| `require` | lista de `id`s que precisam estar instalados antes                                                                   |
| `install` | comando de instalação por sistema: `arc`, `deb`, `mac`, `windows`                                                    |

Se uma lib não tiver comando para o sistema atual, a tentativa de instalar mostra um erro claro.

O app valida os manifestos ao abrir e recusa iniciar se encontrar `id` duplicado, `require` apontando para uma lib inexistente ou dependência circular.

### Regras para os comandos de instalação

Os comandos rodam em segundo plano, sem terminal interativo. Por isso:

- **Devem ser não interativos.** Use `--noconfirm` (pacman), `-y` (apt) e `--accept-package-agreements --accept-source-agreements` (winget).
- **Nunca coloque `| sudo ...`.** O app entrega a senha ao `sudo` pela entrada padrão, e o pipe a ocuparia. Para vários passos como root, use um único `sudo sh -c '...'`.
- **Use `--needed` no pacman**, para não reinstalar o que já existe. Não use `-Sy`, porque sincronizar a base sem atualizar o sistema causa upgrade parcial no Arch/Manjaro.
- **O `check` é um comando só para todos os sistemas.** Para diferenças de nome entre sistemas, use `||` (por exemplo, `python3 --version || python --version`), que funciona tanto no `sh` quanto no `cmd`.

## Estrutura do projeto

```
.
├── main.go            # ponto de entrada: chama cmd.Execute()
├── cmd/               # comandos do Cobra (raiz abre a interface; list)
└── app/
    ├── app.go         # monta system + libs e inicia a interface
    ├── libs/          # modelos Lib/LibStatus e o loader dos manifestos
    │   └── list/      # um arquivo .toml por lib
    ├── system/        # detecção de SO, checks e execução das instalações
    └── tui/           # interface (Bubble Tea): App, delegate da lista, atalhos
```

## Gerando os executáveis

Como o projeto usa apenas Go puro e os manifestos são embutidos, a compilação cruzada só depende de duas variáveis:

```bash
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o dist/devsetup-linux-amd64 .
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o dist/devsetup-darwin-arm64 .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/devsetup-windows-amd64.exe .
```

Para gerar todas as plataformas de uma vez com versão e checksums:

```bash
chmod +x build.sh
VERSION=0.1.0 ./build.sh
```

## Tecnologias

- [Go](https://go.dev)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles) e [Lip Gloss](https://github.com/charmbracelet/lipgloss) para a interface
- [Cobra](https://github.com/spf13/cobra) para a linha de comando
- [BurntSushi/toml](https://github.com/BurntSushi/toml) para os manifestos

## Ideias para o futuro

- `check` por sistema operacional, no mesmo formato do `install`
- instalar automaticamente as dependências (`require`) antes da lib escolhida
- exportar um perfil do que está instalado na máquina atual, para recriar o ambiente depois de formatar
- integração com gerenciadores de versão (como mise ou asdf)

## Licença

A definir.
