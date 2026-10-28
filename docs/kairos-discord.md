# Kairos Discord 협업용 변경

이 포크는 [chenhg5/cc-connect](https://github.com/chenhg5/cc-connect)를 기반으로 합니다. upstream의 기본 동작은 유지하고 아래 두 기능을 설정으로 켤 수 있게 추가합니다.

관리 저장소: [kairos-code-dev/cc-connect](https://github.com/kairos-code-dev/cc-connect). 수정은 이 저장소의 브랜치와 PR로 관리하고, 원본 저장소는 `upstream` remote로 둡니다.

## 봇 간 멘션

Discord 플랫폼 옵션의 `allow_bots_from`에 신뢰하는 상대 봇의 **사용자 ID**를 쉼표로 구분해 적습니다. 역할 ID가 아닙니다. 인간 사용자의 `allow_from`과 별개이며 `*`는 허용하지 않습니다.

- 지정한 봇이 자기 봇을 사용자 멘션 또는 관리 역할 멘션으로 명시적으로 호출할 때만 수신합니다.
- `group_reply_all`을 켜도 봇끼리는 멘션이 필요합니다.
- 자기 메시지, 허용하지 않은 봇, 봇의 전체 멘션만 있는 메시지는 받지 않습니다.
- 모델에는 봇 이름과 ID를 붙여 협업 요청임을 알립니다. 인간 지시로 가장하지 않으며 봇의 본문을 사람의 관리용 slash/shell 명령으로 실행하지 않습니다.
- 이미 같은 채널에서 나눈 대화를 공유하려면 `share_session_in_channel = true`를 지정합니다.

전용 채널에서는 인간 사용자가 멘션 없이 대화하도록 `reply_all_channels`에 해당 채널 ID를 지정합니다. 다른 채널은 기존 멘션 규칙을 유지하며, 다른 봇은 전용 채널에서도 반드시 멘션해야 합니다.

설치된 스킬 이름에 `pages:write-page`처럼 콜론이 있으면 Discord slash 명령 등록에서는 제외합니다. Discord가 이 이름을 거부해 정상 명령까지 일괄 등록 실패하던 것을 방지합니다. 일반 메시지로 스킬을 요청하는 동작에는 영향을 주지 않습니다.

기존 Claude Code 봇과도 이 수신 규칙으로 협업할 수 있습니다. 상대쪽 연결기도 우리 봇을 허용해야 합니다. 메시지를 보고 자동으로 끝없이 답하도록 설정하지 말고, 요청할 때만 상대를 멘션합니다.

## 작업 중 새 입력

Codex 옵션 `backend = "app_server"`, `app_server_url = "stdio://"`, `steer_while_busy = true`를 함께 지정합니다.

지원되는 일반 텍스트는 `turn/steer`와 `expectedTurnId`를 사용해 같은 진행 중인 turn에 전달합니다. 기존 이벤트 수신기를 유지하고 추가 turn을 만들지 않아, 현재 작업이 끝날 때까지 기다리는 것을 줄입니다. 설정을 생략하면 기존 대기열 방식이 유지됩니다.

이 기능은 별도의 대화용 에이전트를 만드는 방식이 아닙니다. Codex가 작업 중 새 질문을 읽고 중간 답변을 생성할 수 있게 입력 경로를 여는 것입니다. 모델의 생성 시간이나 실행 중 도구의 처리 시간까지 없애지는 않으며, 즉시 응답을 보장하지 않습니다. 중간 답변을 먼저 보여주려면 플랫폼의 스트리밍 표시도 켜 두어야 합니다.

다음 경우에는 기존 대기열로 처리합니다.

- 세션 시작 전이거나 active turn이 끝난 경우
- 파일·이미지 첨부가 있어 정상 입력의 파일 저장 과정이 필요한 경우
- 음성·위치 데이터 또는 메시지별 권한 모드 변경이 포함된 경우
- 권한 확인이나 agent 질문의 답변을 기다리는 경우
- 더 먼저 온 메시지가 이미 대기열에 있는 경우
- 해당 backend가 live-input 기능을 지원하지 않거나 설정을 끈 경우

전달 결과를 확인하지 못한 오류는 사용자에게 알리고, 중복 작업을 막기 위해 다른 turn으로 자동 재전송하지 않습니다.

## 설정 예시

`examples/kairos-discord.toml`을 복사해 실제 경로·소유자 ID·상대 봇 ID·봇 토큰을 채웁니다. 실제 토큰 파일은 저장소에 넣지 않습니다. 각 에이전트에는 다른 프로젝트와 다른 Discord 봇 토큰을 사용합니다.

새 Codex 세션을 cc-connect가 관리하도록 시작합니다. 기존 대화를 가져올 필요는 없습니다. 기존 수신기가 같은 봇 토큰으로 실행 중이라면 먼저 그 수신기를 종료하고 봇 점유를 인계합니다.

## 빌드와 원본 업데이트

Go 버전은 `go.mod`를 따릅니다. 먼저 Node.js와 pnpm 10으로 웹 자산을 빌드해야 기본 Go 빌드와 전체 검사를 실행할 수 있습니다.

```sh
git clone --branch fix/discord-bot-live-input https://github.com/kairos-code-dev/cc-connect.git
cd cc-connect
git remote add upstream https://github.com/chenhg5/cc-connect.git
cd web
pnpm install --frozen-lockfile
pnpm build
cd ..
go build -o cc-connect ./cmd/cc-connect
```

Windows에서는 출력 이름을 `cc-connect.exe`로 지정합니다. 바이너리와 실제 설정·토큰은 저장소에 커밋하지 않습니다. 원본 업데이트는 `git fetch upstream` 후 별도 브랜치에서 `upstream/main`을 병합하고, 아래 검사를 거쳐 관리 저장소에 PR로 반영합니다. 이미 공유한 이력에 강제 푸시하지 않습니다.

현재 변경은 PR #1의 시험 브랜치에 있습니다. PR 병합 전에는 위와 같이 브랜치를 지정해야 수정판을 빌드할 수 있습니다.

## 검증

```sh
go test ./core ./agent/codex ./platform/discord -run 'TestCUJ_A10|TestLiveInput_|TestBotPolicy_|TestAppServerSession_OrdinaryLiveInput'
go test ./core/ -run TestCUJ
go test ./...
```

새 회귀 시험은 작업이 끝나지 않은 상태에서 추가 질문 두 개에 대한 답이 플랫폼에 표시되는지, 같은 세션을 유지하는지, 봇 허용 목록과 멘션 제한이 적용되는지 검사합니다. 이 시험들은 모의 모델과 모의 플랫폼을 사용합니다. 실제 Discord와 Codex 모델의 응답 속도는 운영 전 별도 시험해야 합니다.

2026-10-28 Windows 운영 시험에서는 Discord Gateway 연결과 slash 명령 53개 등록을 확인했습니다. 일회성 내부 시험 요청으로 실제 Codex 모델(`gpt-6.1-sol`, effort `low`)을 호출했고, 약 7초의 turn 처리 후 Discord에 계산 답변이 게시된 것을 다시 읽어 확인했습니다. 이는 내부 요청 → 모델 → Discord 전송 시험이며, 실제 사용자 멘션 수신 및 작업 중 추가 입력 시험과는 별개입니다.
