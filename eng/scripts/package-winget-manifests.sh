#!/bin/bash
# Script to create a zip archive of winget manifest files
# This should be run after goreleaser has generated the manifests

set -e

VERSION=${1:-"0.0.1-next"}
PROJECT_NAME="akv"
DIST_DIR="dist"
WINGET_DIR="${DIST_DIR}/winget"
OUTPUT_FILE="${DIST_DIR}/${PROJECT_NAME}-winget-manifests-v${VERSION}.zip"

if [ ! -d "${WINGET_DIR}" ]; then
    echo "Error: Winget directory not found at ${WINGET_DIR}"
    echo "Run 'goreleaser release --snapshot' first to generate manifests"
    exit 1
fi

# Create the zip file
cd "${WINGET_DIR}"
zip -r "../../${OUTPUT_FILE}" .
cd ../..

echo "Created: ${OUTPUT_FILE}"
ls -lh "${OUTPUT_FILE}"
