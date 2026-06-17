## 5. Model and Provider Registry

### 5.1 Provider Config

Provider map key is the provider instance name. `type` selects adapter. `api` selects outbound provider protocol:

```yaml
providers:
  openai:
    type: openai
    api: auto
    api_key_env: OPENAI_API_KEY

  deepseek:
    type: openai-compatible
    api: chat_completions
    base_url: https://api.deepseek.com/v1
    api_key_env: DEEPSEEK_API_KEY
```

### 5.2 Model Aliases

External model names resolve through aliases:

```yaml
models:
  default: default-chat
  aliases:
    default-chat:
      provider: openai
      name: gpt-4.1
    deepseek-chat:
      provider: deepseek
      name: deepseek-chat
```

`GET /v1/models` returns public aliases by default, not all internal provider model names.

### 5.3 Capability Registry

Split:

```text
Provider Registry:
  provider instance -> adapter/config

Model Registry:
  public model alias -> api.ModelRef{Provider, Name}

Capability Registry:
  provider/model -> api.ModelCapabilities
```

Capability source priority:

```text
config override > runtime discovery > built-in defaults
```

`Capability Registry` only answers what is supported. `Assembly` decides downgrade or error.

---

