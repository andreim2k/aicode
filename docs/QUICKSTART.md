# Quick Start Guide

Get up and running with `aicode` in 5 minutes.

## Prerequisites

- macOS or Linux with bash 4.0+
- `jq` installed (`brew install jq` on macOS)
- `claude` CLI installed

## Installation Steps

### 1. Install the Script

```bash
# Copy to your local bin
cp aicode ~/.local/bin/aicode
chmod +x ~/.local/bin/aicode
```

Make sure `~/.local/bin` is in your `$PATH`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
# or ~/.bashrc if using bash
source ~/.zshrc
```

### 2. Create Configuration Directory

```bash
mkdir -p ~/.config/aicode
```

### 3. Add Your First Provider

Copy the example and edit it:

```bash
cp settings.json.example ~/.config/aicode/settings.json
vim ~/.config/aicode/settings.json
```

Add at least one provider with valid credentials.

### 4. Test It

```bash
aicode
```

You should see your providers listed with an interactive menu.

## Basic Usage

```bash
# Start the provider/model selector
aicode

# Use arrow keys to navigate
# Press Enter to expand providers or select models
# Exit with Ctrl+C if needed
```

## Common Setup Scenarios

### Using OpenRouter

1. Sign up at https://openrouter.ai
2. Get API key from settings
3. Update `~/.config/aicode/settings.json`:

```json
[
  {
    "provider": "OpenRouter",
    "base_url": "https://openrouter.ai/api/v1",
    "auth_token": "sk-or-YOUR-KEY-HERE",
    "models": [
      "anthropic/claude-3-5-sonnet",
      "openai/gpt-4-turbo"
    ]
  }
]
```

### Using Local Models with Ollama

1. Install Ollama: https://ollama.ai
2. Run: `ollama serve`
3. Update settings:

```json
[
  {
    "provider": "Local",
    "base_url": "http://localhost:11434/v1",
    "auth_token": "ollama",
    "models": [
      "llama2",
      "mistral"
    ]
  }
]
```

## Troubleshooting

**"command not found: aicode"**
- Make sure `~/.local/bin` is in your PATH
- Try: `source ~/.zshrc` and test again

**"Settings file not found"**
- Check: `cat ~/.config/aicode/settings.json`
- It should exist and contain valid JSON

**Invalid JSON errors**
- Validate your JSON: `jq . ~/.config/aicode/settings.json`
- Use https://jsonlint.com for checking

## Next Steps

- Read [CONFIGURATION.md](CONFIGURATION.md) for advanced setup
- Check the main [README.md](../README.md) for full documentation
- Add multiple providers for different use cases

## Support

If you encounter issues, check:
1. The troubleshooting section in [CONFIGURATION.md](CONFIGURATION.md)
2. Ensure `jq` and `claude` CLI are installed
3. Verify your JSON formatting
