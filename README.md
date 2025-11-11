# aicode

A bash script that provides an interactive menu-driven interface to select and manage AI provider configurations for the Claude CLI.

## Features

- **Interactive Provider Selection**: Navigate providers with arrow keys
- **Expandable Models**: Press Enter on a provider to see its available models indented underneath
- **Dynamic Configuration**: Automatically configures the Claude CLI with selected provider settings
- **Support for Multiple Providers**: Easily manage multiple AI provider configurations
- **Clean Terminal UI**: Color-coded interface with intuitive navigation

## Installation

### Quick Installation (macOS & Linux)

Use the automated installer:

```bash
./install.sh
```

The script will:
- Check system requirements (jq, claude CLI, bash 4+)
- Install aicode to `~/.local/bin`
- Set up configuration directories
- Update your PATH if needed
- Create an example settings file

### Manual Installation

1. Copy the `aicode` script to your local bin directory:
```bash
cp aicode ~/.local/bin/aicode
chmod +x ~/.local/bin/aicode
```

2. Create the configuration directory:
```bash
mkdir -p ~/.config/aicode
```

3. Copy the settings template:
```bash
cp settings.json.example ~/.config/aicode/settings.json
```

4. Edit the settings file with your provider configurations:
```bash
vim ~/.config/aicode/settings.json
```

## Configuration

The script expects a `settings.json` file in `~/.config/aicode/` with the following structure:

```json
[
  {
    "provider": "Provider Name",
    "base_url": "https://api.provider.com/v1",
    "auth_token": "your-auth-token-here",
    "models": [
      "model-1",
      "model-2"
    ]
  }
]
```

### Example Configuration

See `settings.json.example` for a complete example configuration structure.

## Usage

Simply run:
```bash
aicode
```

### Navigation

- **Up/Down Arrows**: Navigate between providers and models
- **Enter**: 
  - On a provider: Expand to show available models
  - On a provider (when expanded): Collapse the models
  - On a model: Select and apply the configuration
- **Left/Right Arrows**: Disabled (no function)

The script will:
1. Display your configured providers
2. Allow you to select a provider and its model
3. Configure the Claude CLI with the selected settings
4. Pass any arguments to the `claude` command

## Environment Requirements

**Supported OS:**
- macOS (10.12 or later)
- Ubuntu/Debian
- Fedora/RHEL/CentOS
- Any Linux with bash 4.0+

**Dependencies:**
- `bash` (4.0+) or `zsh`
- `jq` - JSON query tool
- `claude` CLI installed (from Anthropic)

**Installation help:**
- macOS: `brew install jq`
- Ubuntu/Debian: `sudo apt-get install jq`
- Fedora/RHEL: `sudo dnf install jq`

## File Structure

```
aicode/
├── aicode                  # Main script
├── README.md              # This file
├── settings.json.example  # Example configuration
└── docs/                  # Additional documentation
    └── CONFIGURATION.md   # Detailed configuration guide
```

## Troubleshooting

### Error: "claude command is not installed"
Install the Claude CLI from: https://github.com/anthropics/claude-cli

### Error: "Settings file not found"
Make sure `~/.config/aicode/settings.json` exists and is properly formatted.

### Menu duplication issues
- Clear your terminal and try again
- Ensure your terminal supports ANSI escape sequences

## License

MIT

## Support

For issues and feature requests, please refer to the GitHub repository.
