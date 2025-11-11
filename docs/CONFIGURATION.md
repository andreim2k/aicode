# Configuration Guide

This document explains how to configure the `aicode` script for different AI providers.

## Settings File Location

The settings file must be located at: `~/.config/aicode/settings.json`

## JSON Structure

The settings file is a JSON array containing provider configurations. Each provider object has the following fields:

### Required Fields

- **provider** (string): The display name of your provider (e.g., "OpenRouter", "Local")
- **base_url** (string): The API endpoint URL for the provider
- **auth_token** (string): Your authentication token or API key
- **models** (array): List of available models for this provider

### Example Entry

```json
{
  "provider": "OpenRouter",
  "base_url": "https://openrouter.ai/api/v1",
  "auth_token": "sk-or-xxx-xxxxxxxxxxxx",
  "models": [
    "anthropic/claude-3-5-sonnet",
    "anthropic/claude-3-opus"
  ]
}
```

## Setting Up Different Providers

### OpenRouter

1. Sign up at https://openrouter.ai
2. Get your API key from the dashboard
3. Add to `settings.json`:

```json
{
  "provider": "OpenRouter",
  "base_url": "https://openrouter.ai/api/v1",
  "auth_token": "sk-or-YOUR-KEY-HERE",
  "models": [
    "anthropic/claude-3-5-sonnet",
    "anthropic/claude-3-opus",
    "openai/gpt-4-turbo"
  ]
}
```

### Local LLM Server

If you're running a local LLM server (e.g., Ollama, LM Studio):

```json
{
  "provider": "Local",
  "base_url": "http://localhost:8000/v1",
  "auth_token": "local-api-key",
  "models": [
    "llama-2-70b",
    "mistral-large"
  ]
}
```

### Anthropic Direct (if using their API)

```json
{
  "provider": "Anthropic",
  "base_url": "https://api.anthropic.com/v1",
  "auth_token": "sk-ant-YOUR-KEY-HERE",
  "models": [
    "claude-3-5-sonnet",
    "claude-3-opus",
    "claude-3-haiku"
  ]
}
```

## Security Notes

- **Never commit** your actual `settings.json` to version control
- Use `.gitignore` to exclude it from the repository
- Keep your API keys secure and rotate them regularly
- If accidentally committed, regenerate your API keys immediately

## Validation

The script validates your `settings.json` file using `jq`. Make sure:

1. The JSON is properly formatted (use a JSON validator if unsure)
2. All required fields are present
3. The `auth_token` field exists (even if empty initially)
4. The `models` array is not empty

## Testing Your Configuration

After setting up your `settings.json`, test it:

```bash
# Run the script
aicode

# Navigate and select a provider and model
# The script will show you the selected configuration
```

## Troubleshooting Configuration Issues

### "Invalid JSON in settings.json"

- Check your JSON formatting at: https://jsonlint.com/
- Ensure all strings are properly quoted
- Verify commas are in the right places

### "Provider configuration not found"

- Make sure the provider name in the menu matches what's in `settings.json`
- Check that the JSON structure follows the required format

### "Error: Provider configuration is incomplete"

- Ensure `base_url` and `auth_token` fields are present
- They can be empty strings but must exist in the JSON

## Advanced Configuration

### Multiple Profiles

You can have multiple provider configurations for different use cases:

```json
[
  {
    "provider": "Development",
    "base_url": "http://localhost:8000/v1",
    "auth_token": "dev-key",
    "models": ["llama-2-70b"]
  },
  {
    "provider": "Production",
    "base_url": "https://api.provider.com/v1",
    "auth_token": "prod-key",
    "models": ["claude-3-opus", "gpt-4"]
  }
]
```

### Model Organization

Group related models by provider:

```json
{
  "provider": "Multi-Model Provider",
  "base_url": "https://api.example.com/v1",
  "auth_token": "xxx",
  "models": [
    "fast-model",
    "accurate-model",
    "balanced-model"
  ]
}
```

## Updating Configuration

To add or modify providers:

1. Edit `~/.config/aicode/settings.json`
2. Make your changes
3. Save the file
4. Run `aicode` again - it will validate and use the new configuration

Changes take effect immediately without restarting the script.
