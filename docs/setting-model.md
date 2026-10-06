# 設定本機模型 (Windows + Ollama + podman)

在 Windows 上用 Ollama 提供 LLM 與 embedding 模型,給跑在 podman machine (WSL) 裡的 WeKnora 使用.

## 1. 安裝 Ollama 並下載模型

```powershell
winget install Ollama.Ollama
ollama pull gemma3:12b        # LLM,約 8GB,需 12GB 以上 VRAM
ollama pull embeddinggemma    # embedding,約 600MB,768 維
```

- VRAM 較小可改用 `gemma3:4b`.
- 重視中文檢索品質可改用 `bge-m3` (約 1.2GB,1024 維).
- **embedding model 選定後盡量不要換**,換了要重建整個知識庫的索引.

## 2. 讓 Ollama 接受 WSL 連線

container 裡的 `host.docker.internal` 指向的是 podman machine 這台 WSL VM,不是 Windows,所以要改用 Windows 在 WSL 網路上的 IP.

**a. 讓 Ollama 監聽所有網卡**,設定後從系統匣結束 Ollama 再重新開啟:

```powershell
[Environment]::SetEnvironmentVariable('OLLAMA_HOST', '0.0.0.0', 'User')
```

**b. 開放防火牆,只允許 WSL 網段** (以系統管理員身分執行 PowerShell):

```powershell
New-NetFirewallRule -DisplayName "Ollama (WSL)" -Direction Inbound -Protocol TCP -LocalPort 11434 -RemoteAddress 172.16.0.0/12 -Action Allow
```

**c. 查詢 Windows 在 WSL 網路上的 IP:**

```powershell
Get-NetIPAddress -AddressFamily IPv4 | Where-Object InterfaceAlias -match 'WSL' | Select-Object IPAddress
```

## 3. 設定 WeKnora

把 `.env` 的 `OLLAMA_BASE_URL` 改成上一步查到的 IP,例如:

```
OLLAMA_BASE_URL=http://172.26.16.1:11434
```

重建 app 並測試連線,有回傳 JSON 就代表成功:

```powershell
podman compose up -d app
podman exec knowledge-hub-app curl -s http://172.26.16.1:11434/api/tags
```

## 4. 在 UI 設定模型

在模型設定中選 **Ollama**:

| 類型 | 模型名稱 |
|---|---|
| LLM | `gemma3:12b` |
| Embedding | `embeddinggemma` |

Ollama 位址一律寫在 `.env`.不要在 UI 的 Base URL 欄位填私有 IP,UI 的輸入會經過 SSRF 檢查而被擋下.

## 疑難排解

- **重開機後連不到 Ollama**:WSL NAT 的 IP 可能變了.重新執行步驟 2c,更新 `.env` 後再 `podman compose up -d app`.
- **想固定位址**:在 `%UserProfile%\.wslconfig` 的 `[wsl2]` 區段加上 `networkingMode=mirrored`,再執行 `wsl --shutdown`.這樣 WSL 和 Windows 會共用 localhost,但會影響所有 WSL distro.此做法尚未在本專案環境驗證過,切換後要用上面的 `curl` 指令確認連線.
- **改了 `.env` 沒生效**:要用 `podman compose up -d`,`restart` 不會重新讀取 `.env`.
