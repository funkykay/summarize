#!/usr/bin/env bash

set -euo pipefail

APP_NAME="summarize"
BIN_DIR="${HOME}/bin"
BIN_PATH="${BIN_DIR}/${APP_NAME}"
TMP_DIR="$(mktemp -d)"
REPO_SLUG="${GITHUB_REPOSITORY:-funkykay/summarize}"

cleanup() {
  rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

fail() {
  echo "Error: $*" >&2
  exit 1
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    fail "Required command not found: $1"
  fi
}

detect_repo_slug() {
  if [[ -n "${REPO_SLUG}" ]]; then
    return
  fi

  if command -v git >/dev/null 2>&1; then
    local remote_url
    remote_url="$(git config --get remote.origin.url 2>/dev/null || true)"

    if [[ "${remote_url}" =~ ^git@github\.com:(.+/.+)\.git$ ]]; then
      REPO_SLUG="${BASH_REMATCH[1]}"
      return
    fi

    if [[ "${remote_url}" =~ ^https://github\.com/(.+/.+)\.git$ ]]; then
      REPO_SLUG="${BASH_REMATCH[1]}"
      return
    fi

    if [[ "${remote_url}" =~ ^https://github\.com/(.+/.+)$ ]]; then
      REPO_SLUG="${BASH_REMATCH[1]}"
      return
    fi
  fi

  fail "Unable to determine GitHub repository. Set GITHUB_REPOSITORY=owner/repo."
}

fetch_latest_tag() {
  local effective_url
  effective_url="$(curl -fsSL -o /dev/null -w "%{url_effective}" "https://github.com/${REPO_SLUG}/releases/latest")" \
    || return 1

  local latest_tag
  latest_tag="${effective_url##*/}"

  if [[ -z "${latest_tag}" || "${latest_tag}" == "latest" ]]; then
    return 1
  fi

  printf '%s\n' "${latest_tag}"
}

normalize_version() {
  local version="$1"
  version="${version#v}"
  printf '%s\n' "${version}"
}

get_installed_version() {
  if [[ ! -x "${BIN_PATH}" ]]; then
    return 1
  fi

  local raw_version
  raw_version="$("${BIN_PATH}" version 2>/dev/null || true)"
  if [[ -z "${raw_version}" ]]; then
    return 1
  fi

  printf '%s\n' "${raw_version##* }"
}

confirm_update() {
  local current_version="$1"
  local target_version="$2"

  printf 'Would you like to update from %s to %s? [y/N] ' "${current_version}" "${target_version}" >&2
  read -r answer

  case "${answer}" in
    y|Y|yes|YES)
      return 0
      ;;
    *)
      echo "Aborted."
      return 1
      ;;
  esac
}

detect_platform_asset_name() {
  local os
  local arch

  os="$(uname -s)"
  arch="$(uname -m)"

  case "${os}" in
    Linux)
      case "${arch}" in
        x86_64|amd64)
          printf '%s\n' "summarize-linux-x64"
          return 0
          ;;
        aarch64|arm64)
          printf '%s\n' "summarize-linux-arm64"
          return 0
          ;;
      esac
      ;;
    Darwin)
      case "${arch}" in
        arm64)
          printf '%s\n' "summarize-macos-arm64"
          return 0
          ;;
      esac
      ;;
  esac

  return 1
}

install_binary_release() {
  local download_url="$1"

  mkdir -p "${BIN_DIR}"

  curl -fsSL "${download_url}" -o "${TMP_DIR}/${APP_NAME}"
  install -m 0755 "${TMP_DIR}/${APP_NAME}" "${BIN_PATH}"
}

main() {
  require_command curl

  detect_repo_slug

  local latest_tag
  latest_tag="$(fetch_latest_tag)" \
    || fail "Unable to fetch latest GitHub release for ${REPO_SLUG}."

  local target_version
  target_version="$(normalize_version "${latest_tag}")"

  local current_version=""
  if current_version="$(get_installed_version)"; then
    current_version="$(normalize_version "${current_version}")"

    if [[ "${current_version}" == "${target_version}" ]]; then
      echo "${APP_NAME} ${target_version} is already installed at ${BIN_PATH}."
      exit 0
    fi

    confirm_update "${current_version}" "${target_version}" || exit 0
  fi

  local asset_name
  asset_name="$(detect_platform_asset_name)" \
    || fail "No standalone release asset available for platform $(uname -s)/$(uname -m)."

  local download_url
  download_url="https://github.com/${REPO_SLUG}/releases/download/${latest_tag}/${asset_name}"

  install_binary_release "${download_url}" \
    || fail "Unable to download release asset '${asset_name}' from ${latest_tag}."

  local installed_version
  installed_version="$(get_installed_version || true)"
  if [[ -n "${installed_version}" ]]; then
    echo "Installed ${APP_NAME} ${installed_version} to ${BIN_PATH}."
    exit 0
  fi

  echo "Installed ${APP_NAME} to ${BIN_PATH}."
}

main "$@"
