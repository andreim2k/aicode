#!/bin/bash

# Z.AI Adapter - Converts Anthropic API format to Z.AI OpenAI-compatible format
# This script acts as a bridge between Claude CLI (Anthropic API) and Z.AI (OpenAI API)

set -euo pipefail

# Get configuration from environment or .claude/settings.json
if [ -f "${HOME}/.claude/settings.json" ]; then
    BASE_URL=$(jq -r '.env.ANTHROPIC_BASE_URL // empty' "${HOME}/.claude/settings.json")
    AUTH_TOKEN=$(jq -r '.env.ANTHROPIC_AUTH_TOKEN // empty' "${HOME}/.claude/settings.json")
    MODEL=$(jq -r '.env.ANTHROPIC_MODEL // empty' "${HOME}/.claude/settings.json")
fi

# Fall back to environment variables
BASE_URL="${ANTHROPIC_BASE_URL:-${BASE_URL:-}}"
AUTH_TOKEN="${ANTHROPIC_AUTH_TOKEN:-${AUTH_TOKEN:-}}"
MODEL="${ANTHROPIC_MODEL:-${MODEL:-}}"

if [ -z "$BASE_URL" ] || [ -z "$AUTH_TOKEN" ] || [ -z "$MODEL" ]; then
    echo "Error: Missing Z.AI configuration" >&2
    echo "Please ensure ANTHROPIC_BASE_URL, ANTHROPIC_AUTH_TOKEN, and ANTHROPIC_MODEL are set" >&2
    exit 1
fi

# Read the Anthropic API request from stdin
read_request() {
    local content=""
    while IFS= read -r line; do
        content="${content}${line}"
    done
    echo "$content"
}

# Convert Anthropic messages format to OpenAI format
convert_messages() {
    local anthropic_json="$1"
    
    # Extract messages array and convert role format
    echo "$anthropic_json" | jq '
        .messages |= map(
            if .role == "user" then .
            elif .role == "assistant" then .
            else . 
            end
        )
    '
}

# Handle /v1/messages endpoint - convert to /chat/completions
handle_messages_request() {
    local request_body="$1"
    
    # Convert Anthropic format to OpenAI format
    local openai_request=$(echo "$request_body" | jq '{
        model: .model,
        messages: .messages,
        max_tokens: .max_tokens,
        temperature: .temperature,
        top_p: .top_p,
        system: .system
    } | with_entries(select(.value != null))')
    
    # Make request to Z.AI chat/completions endpoint
    local response=$(curl -s \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "$openai_request" \
        "${BASE_URL%/}/chat/completions")
    
    # Convert OpenAI response to Anthropic format
    convert_openai_to_anthropic_response "$response"
}

# Convert OpenAI response format to Anthropic format
convert_openai_to_anthropic_response() {
    local openai_response="$1"
    
    # Check if response has error
    if echo "$openai_response" | jq -e '.error' >/dev/null 2>&1; then
        echo "$openai_response"
        return
    fi
    
    # Convert OpenAI chat completion response to Anthropic message response
    echo "$openai_response" | jq '{
        id: "msg_" + (.id | gsub("-"; "_")),
        type: "message",
        role: "assistant",
        content: [
            {
                type: "text",
                text: .choices[0].message.content
            }
        ],
        model: .model,
        stop_reason: (if .choices[0].finish_reason == "stop" then "end_turn" else .choices[0].finish_reason end),
        stop_sequence: null,
        usage: {
            input_tokens: .usage.prompt_tokens,
            output_tokens: .usage.completion_tokens,
            cache_creation_input_tokens: 0,
            cache_read_input_tokens: 0
        }
    }'
}

# Main - determine which endpoint is being called
main() {
    # Claude SDK reads from environment and makes requests
    # We need to intercept the API call
    # Since we can't easily intercept HTTP calls, we'll provide a modified base URL behavior
    
    # For now, just read the request and handle it
    local request_body=$(read_request)
    
    # Check if this looks like a messages request
    if echo "$request_body" | jq -e '.messages' >/dev/null 2>&1; then
        handle_messages_request "$request_body"
    else
        echo "$request_body"
    fi
}

# Run if stdin has data, otherwise just provide functions
if [ -t 0 ]; then
    # Interactive terminal - no input
    echo "Z.AI Adapter ready. Use via ANTHROPIC_BASE_URL environment override." >&2
else
    # Data from pipe
    main
fi
