# 中國相關整合清除與上游更新作業手冊

本文件是 Knowledge Hub fork 的維護規範。每次吸收上游程式、更新依賴、產生文件或發行映像時，都要依此重新審查，避免已移除的中國營運服務、連線端點與品牌內容被帶回來。**不能只搜尋品牌名稱就認定沒有外連風險；也不能因為原始碼有封鎖清單就宣稱所有外連都已被阻止。**

目前的程式清理仍在進行中。下方「保留與未完成項目」列出尚未達到嚴格「所有中國來源程式碼都移除」標準的項目；更新上游時不得把它們誤記為已解決。

## 1. 範圍與判斷原則

| 類別 | 本 fork 的處理方式 | 審查重點 |
| --- | --- | --- |
| 中國營運的雲端模型、搜尋、通訊、儲存、文件解析與託管服務 | 移除註冊、實作、預設設定、前端選項、金鑰表單、說明和發行流程 | 不得再有可選用的服務或自動連線路徑 |
| 中國服務的舊資料列與已存金鑰 | 保留必要的資料庫相容結構，但執行路徑必須拒絕已移除的 provider 與端點 | 尤其檢查通用 OpenAI 相容 provider 的回退行為 |
| 自行部署的 Milvus、Apache Doris | 依使用者決定保留 | 不得暗中改用供應商託管端點；檢查預設位址與映像來源 |
| 顯示品牌與簡體中文介面／文件 | 顯示 Knowledge Hub；移除 zh-CN 介面與舊品牌宣傳內容 | 新增功能須提供現有語系翻譯；檢查硬編碼文字與圖片資產 |
| 歷史技術識別字 | 必要時保留 fork Go module、環境變數、API、資料表或檔名以維持升級相容 | 保留須有具體原因，不能因此恢復舊服務 |
| 著作權、授權與第三方聲明 | 保留真實權利人與原授權文字 | 不得為了品牌清理而刪除法定歸屬資訊 |

此政策針對服務行為、預設外連及產品內容。供應商來源地本身不能替代依賴審查；例外依賴仍須記錄其用途、連線行為、授權與移除計畫。

## 2. 已實施的基線

以下是目前工作樹的目標狀態，不代表日後上游的檔案布局不會改變。更新時應同時檢查註冊點、型別、生成資料、UI、設定、文件和測試。

| 功能面 | 允許／保留 | 已移除、不得恢復的例子 | 主要檢查位置 |
| --- | --- | --- | --- |
| 模型 provider | Anthropic、Azure OpenAI、Gemini、Generic、GPUStack、Jina、LiteLLM、Novita、NVIDIA、OpenAI、OpenRouter、Requesty | Aliyun／DashScope、DeepSeek、Hunyuan／LKEAP、LongCat、MiMo、MiniMax、ModelScope、Moonshot、Qianfan、Qiniu、SiliconFlow、Volcengine／Doubao、WeKnoraCloud、Zhipu 等 | `internal/models/providers/builtin.go`、`internal/models/catalog/data/`、`internal/models/runtime/resolve.go`、`internal/models/parity/catalog_test.go`、`frontend/src/` |
| 模型協定與端點 | 上列 provider 所需協定、Ollama 與明確核准的自訂端點 | Tencent LKEAP、Ark、DashScope 特有客戶端及中國預設 API URL | `internal/models/api/`、`internal/models/chat/`、`embedding/`、`rerank/`、`asr/`、`vlm/`、`config/builtin_models.yaml.example` |
| 網頁搜尋 | DuckDuckGo、Google、Bing、Tavily、Ollama、SearXNG、Keenable、Exa、Brave、Serply | Baidu、Bocha、Metaso、Zhipu | `internal/container/container.go` 的註冊、`internal/types/web_search_provider.go`、`internal/infrastructure/web_search/`、前端設定頁 |
| 即時通訊 | Slack、Telegram、Mattermost | DingTalk、Feishu／Lark、QQ Bot、WeChat、WeCom、Yunzhijia | `internal/handler/im.go`、`internal/im/`、前端通道選項、設定及文件 |
| 資料來源 connector | Confluence、GitLab、Notion、RSS | DingTalk、Feishu／Lark(含 Drive)、Tencent IMA、Yuque | `internal/datasource/connector/`、`internal/datasource/connector.go`、`frontend/src/views/knowledge/settings/DataSourceEditorDialog.vue`、datasource i18n |
| 物件儲存 | local、MinIO、S3 相容儲存;Compose 的 MinIO 映像為 `cgr.dev/chainguard/minio`(可用 `MINIO_IMAGE` 覆寫) | COS、TOS、OSS、KS3、OBS 的專用實作與設定;上游改用的 `pgsty/minio` 社群映像 | `internal/application/service/file/`、`internal/storageallowlist/`、`.env*`、`docker-compose*.yml`、`helm/values.yaml` |
| 向量儲存 | PostgreSQL／pgvector、Elasticsearch、OpenSearch、Qdrant、Weaviate，以及自架 Milvus、Doris | Tencent VectorDB | `internal/types/retriever.go`、`internal/types/vectorstore.go`、`internal/application/repository/retriever/`、Compose／Helm |
| 文件解析 | DocReader builtin、simple、anydoc | MinerU、PaddleOCR-VL、WeKnoraCloud reader | `internal/infrastructure/docparser/engines.go`、`docreader/`、前端引擎選項、範例環境檔 |
| 沙箱與技能 | 目前保留的本地、Docker、E2B 等實作；技能來源須經驗證 | CubeSandbox 託管整合、SkillHub.cn 下載來源 | `internal/sandbox/`、`internal/application/service/tenant_skill_source.go`、`scripts/` |
| 發行與文件 | `knowledge-hub-*` 映像、英文站點與 API 文件、fork GitHub 路徑 | 原品牌的宣傳站、zh-CN locale、自動發佈舊 MCP PyPI 套件、鏡像站預設來源 | `.github/workflows/`、`frontend/src/i18n/`、`website-docs/`、`mcp-server/`、lockfiles |

模型執行時的拒絕清單位於 `internal/models/runtime/resolve.go`，測試位於 `resolve_policy_test.go`。它應在通用 provider 回退之前檢查舊 provider ID 與 URL 主機。這是防止舊資料列把已存金鑰送往舊端點的保護層；新增上游 provider 時，仍要審查所有建立客戶端和測試連線的路徑。

`internal/application/service/tenant_skill_source.go` 對 SkillHub.cn 主機的明確拒絕、`internal/utils/security.go` 對雲端 metadata 主機的阻擋，以及 `scripts/sanitize_swagger.py` 對舊 schema 的清理，**會故意包含被禁名稱**。搜尋命中必須看上下文，不能刪除這些防護。

## 3. 保留與未完成項目

1. **Go module** 固定為 `github.com/ai-tool-collection/WeKnora`。變更它會影響大量 import、生成碼與使用者現有依賴；顯示品牌則使用 Knowledge Hub。`WEKNORA_*` 環境變數、部分 CLI／API 識別字、Redis key、資料庫欄位與歷史遷移也可能為相容性保留。新增舊識別字必須說明其必要性。
2. **Milvus 與 Apache Doris** 是使用者明確核准的自架開源後端。不要因名稱來源而從更新中刪除；也不要把它們改接未審核的託管服務。
3. **TDesign** 元件與圖示套件目前仍在前端依賴中。圖示 sprite 由本地提供，`frontend/src/utils/tdesign-icon-offline.ts` 用來阻止套件嘗試載入其 CDN；正式產物可能仍含供應商 URL 字串。是否完整替換 TDesign 尚未決定，故嚴格「沒有中國來源依賴」檢查目前**不通過**。套件升級須重新驗證瀏覽器實際網路請求，不能只信任既有阻擋邏輯。
4. **BrowserSkill** 是可選的本地整合。`scripts/build_browserskill.sh` 只接受經審查、符合 `scripts/browserskill-release.json` 指定 commit 的本地原始碼，不再自動 clone；其授權聲明仍須隨資產保存。若政策改成連這個來源也不得保留，必須一併移除功能、構建、UI、設定與文件。
5. `go.mod` 仍有 ByteDance Sonic 相關的**間接依賴**，需追查引入鏈並評估替代方案。請勿把「服務已移除」寫成「供應鏈已完全排除中國來源」。
6. 後端仍有歷史簡體中文註解、錯誤訊息、IM 文案與測試資料；部分發行檔名、腳本與範例仍有技術性舊品牌字串。清理尚未完成。翻譯使用者可見字串時，要同步更新測試及所有既有語系，並避免改動需要相容的儲存值。
7. **MinIO 映像**:quay.io 自 2026-09-24 關閉 `minio/minio` 匿名拉取,上游改用 Pigsty 維護的 `pgsty/minio`;本 fork 改用 Chainguard 從原始碼建置的 `cgr.dev/chainguard/minio`。其免費版只提供 `latest` tag,需要固定版本時以 digest 釘選。映像預設以 uid 65532 執行且沒有 `curl`,Compose 因此設 `user: "0:0"` 相容舊版 root 寫入的 volume,healthcheck 改用 `mc ready local`。
8. **Swagger 產物**:已提交的 `docs/swagger.*` 仍含已移除的 `/wechat/qrcode` 路由與 COS、TOS、OSS、KS3、OBS、Cube 定義,且缺少 `/wiki-search` 等新端點。以 swag v1.16.4/v1.16.6 執行 `make docs` 會改變所有 definition 命名並遺失 `ErrorCode` enum,導致 `go test ./docs` 失敗;需先確認上游使用的 swag 版本再重新生成。
9. 來源與 lockfile 掃描不能證明執行時絕無外連。自訂模型 URL、插件、MCP、網頁擷取與使用者設定的 S3 端點仍可產生對外流量；部署者應以 egress allowlist 和網路紀錄驗證實際目的地。

## 4. 每次更新上游的標準流程

### A. 準備與差異盤點

1. 先把此 fork 的清理工作提交成可回溯的基線，確認 `git status --short` 為空。
2. 記下本 fork 與上游的 commit SHA、版本與預計納入的功能。`git remote -v` 確認來源；目前只有 `origin`，沒有預先設定的 `upstream`。上游由維護者同步到 `main`,再以 `git merge main` 帶入清理分支。
3. 在獨立分支／worktree 作業。先用 `git diff --name-status HEAD..<reviewed-upstream-ref>` 與 `git log --oneline HEAD..<reviewed-upstream-ref>` 檢視變更，記錄新增的資料夾、依賴、workflow、環境變數、外部 URL 和 UI 選項。
4. 合併或挑選上游功能時，逐項帶入需要的功能與安全修正；不要用整個上游目錄覆蓋此 fork 的 provider、IM、儲存、解析、部署或文件目錄。

### B. 復原清理邊界

依第 2 節的矩陣逐層檢查，尤其是以下容易只改一半的鏈條：

1. **後端能力鏈：**實作 → 註冊／工廠 → 型別與列舉 → handler／路由 → 設定與環境變數 → 資料遷移 → 測試。移除實作後，不能讓通用 provider 對舊 ID 自動回退。
2. **前端能力鏈：**型別 → 選單與表單 → 圖示資產 → i18n → 範例與測試。後端拒絕的服務不得仍顯示為可設定項。
3. **發行鏈：**Dockerfile、Compose、Helm、GitHub Actions、MCP package metadata、lockfile、更新檢查器與範例 URL。檢查是否引入新的中國映像／套件 registry 或自動下載腳本。
4. **生成鏈：**模型 catalog seed／generated JSON、Swagger、protobuf、前端 widget 與文件站。修改來源後重新生成，再檢查生成產物；不要只手改生成檔。protobuf 生成檔(`*.pb.go`、`*_pb2.py`)的序列化 descriptor 帶長度前綴,直接文字取代 `go_package` 會讓程式在 init 時 panic。
5. **升級資料：**對已存在的 provider、通道與儲存設定，確認舊金鑰不會送往被移除服務。歷史欄位與遷移應以向後相容為優先；停用執行路徑與安全移轉使用者資料可分開處理。
6. **Go module path:**上游 import 為 `github.com/Tencent/WeKnora`,合併後新增與衝突檔案都要改寫為 fork module path,否則無法編譯。
7. **被刪除功能的新增檔案:**上游對已移除功能新增的檔案(例如 connector 測試)不會產生衝突,會被靜默帶入;要用 `git diff --name-only --diff-filter=A` 另行檢查。
8. **共用 helper 與測試:**刪除整合檔案前確認其中沒有被保留功能共用的函式(例如 E2B 沙箱曾依賴 Cube 檔案內的 `parseProxyURL`),並同步修改引用已移除型別的測試;以具 CGO 的 `go vet ./...` 與 `go test ./...` 驗證,不能只編譯主程式。
9. **文件站與 Swagger 衝突:**`website-docs/` 是 fork 自行維護的英文精簡版,上游刪除或改寫的中文頁面不直接合併;fork 已刪除的頁面維持刪除,仍存在的頁面只把行為變更改寫成英文。`docs/swagger.yaml` 在 fork 中與 `swagger.json` 內容相同(JSON 也是合法 YAML),衝突時先解 `swagger.json`,再複製成 `swagger.yaml`;`docs/docs.go` 保留 `//go:embed swagger.json`,不要換回上游的內嵌字串模板。
10. **沙箱映像:**上游 workflow、`scripts/build_images.sh` 與 `docker/Dockerfile.sandbox` 的 `cube`/`desktop-cube` target 與 `cubesandbox-base` 映像一律不帶入;只吸收對 `runtime`/`desktop` 通用的建置修正。

### C. 靜態搜尋與人工分類

以下命令從專案根目錄執行。命中可能是拒絕清單、測試、授權聲明、相容識別字或本文件；**逐筆分類**為「實際整合」、「防護／回歸測試」、「歷史相容」、「法律聲明」、「待清理文字」。任何無法解釋的新增命中都阻止合併。

```sh
rg --hidden -n -i 'aliyun|alibaba|dashscope|deepseek|doubao|hunyuan|lkeap|longcat|mimo|minimax|modelscope|moonshot|qianfan|qiniu|siliconflow|volcengine|weknoracloud|zhipu' internal frontend/src config docreader scripts .github helm
rg --hidden -n -i 'baidu|bocha|metaso|dingtalk|feishu|larksuite|qqbot|wechat|wecom|yunzhijia|yuque|ima\.qq|pgsty|tencent.?vectordb|mineru|paddleocr|cube.?sandbox|skillhub\.cn' internal frontend/src config docreader scripts .github helm docker-compose.yml docker-compose.dev.yml
rg --hidden -n -i 'aliyuncs\.com|baidubce\.com|tencentcloudapi\.com|myqcloud\.com|qcloud\.com|volces\.com|bigmodel\.cn|modelscope\.cn|siliconflow\.cn|weixin\.qq\.com|dingtalk\.com' internal frontend/src config docreader scripts .github helm
rg --hidden -n -i 'cos|tos|oss|zh-CN|zh_CN|zh-Hans|WeKnora|Tencent' .env.example .env.lite.example docker-compose.yml docker-compose.dev.yml helm frontend/src website-docs mcp-server Makefile SECURITY.md
rg --hidden -n -i 'registry\.npm\.taobao\.org|npmmirror\.com|registry\.nlark\.com|mirrors\.aliyun|hub\.docker\.com/r/' frontend/package-lock.json website-docs/package-lock.json go.mod go.sum mcp-server/uv.lock
```

`cos`、`tos`、`oss` 及 `Tencent` 會產生許多非服務命中；它們用於人工複核，不能直接當成零命中 gate。也要額外看 `git diff --name-only` 中新加入的 binary、SVG、HTML、`.env`、workflow、lockfile 和生成資料，因為純文字搜尋未必能涵蓋它們。

### D. 可重複的驗證命令

需在具有相應工具鏈的環境執行；任何跳過或失敗必須記錄原因與替代驗證。根目錄 Go 後端使用 `pg_query`，須有可用的 CGO C 編譯器。Windows 若只能 `CGO_ENABLED=0`，`pg_query.Parse/Deparse` 會缺失，這**不是**測試通過。

```sh
# 專案根目錄：模型封鎖、provider 名單、後端與 API 文件
go test ./internal/models/runtime ./internal/models/parity ./docs
go test ./...
make model-catalog-check
make docs  # runs swag and then scripts/sanitize_swagger.py
go test ./docs

# 前端：請使用 package.json 的實際 script 名稱
cd frontend
npm ci
npm run check-i18n
npm run type-check
npm test
npm run build
cd ..

# 文件站
cd website-docs
npm ci
npm run build
cd ..

# 獨立 Go module
cd cli && go test ./... && cd ..
cd client && go test ./... && cd ..

git diff --check
```

`docs/swagger.json` 與 `docs/swagger.yaml` 應保持有效、沒有已移除路由和未解析 `$ref`；`scripts/sanitize_swagger.py` 只處理生成結果，不能替代修正來源註解或 handler。模型 catalog 的 provider 清單必須與 `internal/models/parity/catalog_test.go` 相符；新增允許 provider 時同時更新測試與本文件。

### E. 執行時煙霧測試與外連驗證

1. 在隔離的 staging 環境，以全新資料庫啟動 Compose／Helm；確認登入、建立空間、檔案上傳與解析、建立知識庫、檢索、模型測試連線、聊天及引用、MCP、保留的 IM 通道可運作。
2. 用升級前資料庫的**去識別化副本**測試升級。對已移除 provider 與端點的舊資料列，確認 UI 不提供它們，後端回傳不支援錯誤，且沒有送出任何帶金鑰的請求。
3. 分別測試本地檔案、MinIO／S3、自架 Milvus／Doris；不要只跑預設 PostgreSQL 向量路徑。使用者沒有啟用的整合不應阻止核心功能啟動。
4. 在瀏覽器 Network、容器 DNS／proxy／防火牆日誌中檢查啟動、登入、設定頁、圖示、搜尋、文件解析和模型測試連線。審查所有意外目的地；若 TDesign 仍保留，必須確認沒有對其 CDN 的實際請求。
5. 以 egress allowlist 限制部署，只放行明確核准的模型、搜尋、OIDC、MCP、S3 等目的地。不要把 `.cn` 頂級域名掃描當成完整邊界；供應商可能使用 `.com`，使用者設定也可能指向任意主機。

## 5. 合併驗收紀錄範本

每次上游更新在 PR 描述或變更紀錄中填寫：

```text
上游來源／commit：
本 fork 基線／commit：
納入功能與安全修正：
新增／變動的外部服務、URL、套件、映像、workflow：
已排除的中國整合與修改位置：
保留的相容識別字及理由：
TDesign／BrowserSkill／間接依賴狀態：
舊資料列與金鑰的拒絕測試：
靜態搜尋新增命中及逐項分類：
Go、前端、文件、CLI、client 測試結果與跳過原因：
staging 功能與實際外連驗證結果：
本文件的新增規則或例外：
```

**完成條件：**沒有未審查的中國服務執行路徑或預設外連；保留功能的測試與煙霧測試通過；既有金鑰不會被舊 provider 回退送出；所有命中與例外有紀錄；本文與程式同步更新。若嚴格要求移除一切中國來源依賴，還必須先解決第 3 節列出的 TDesign、BrowserSkill 與間接依賴，不能以本手冊的其他檢查通過代替。
