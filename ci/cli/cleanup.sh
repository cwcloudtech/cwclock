#!/usr/bin/env bash
set -e

source ./ci/cli/compute-env.sh
cd "${CWCLOCK_CLI_DIR}"

MAX_RELEASES=5
API_URL="${GITLAB_HOST}/api/v4/projects/${GITLAB_PROJECT}"

gitlab_api_call() {
    local method=$1
    local endpoint=$2
    local response
    
    response=$(curl -s -X "$method" \
        --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
        "${API_URL}${endpoint}")
    
    if [[ $? -ne 0 ]]; then
        echo "Failed to make API call: $endpoint"
        exit 1
    fi
    
    echo "$response"
}

#? Get releases sorted by created_at
releases=$(gitlab_api_call "GET" "/releases?per_page=100&sort=desc" | \
    jq -r '.[] | [.created_at, .tag_name] | @csv' | \
    sort -r)

#? Count releases and delete older ones
count=0
while IFS=, read -r created_at tag_name; do
    count=$((count + 1))
    
    if [ "$count" -gt "$MAX_RELEASES" ]; then
        # Remove quotes from tag_name
        tag_name=$(echo "$tag_name" | tr -d '"')
        echo "Deleting release: $tag_name"
        
        gitlab_api_call "DELETE" "/releases/$tag_name"
        if [[ $? -eq 0 ]]; then
            echo "Successfully deleted release: $tag_name"
        else
            echo "Failed to delete release: $tag_name"
            exit 1
        fi
    fi
done <<< "$releases"

echo "Cleanup completed successfully"
