#!/usr/bin/env uv run
#/// script
# requirements: ["requests", "tqdm"]
# ///

# This script audits repositories for the presence and correctness of the integration manifest.
# A repository can be in the following states:
# - No manifest: Not an integration repo (or missing manifest)
# - Manifest present but an older version
# - Manifest present but invalid: Needs correction
# - Manifest present and valid: Correctly configured integration repo
#
# Manifest present but invalid = needs correction (report)
# manifest present but older version = needs correction (report)
# No manifest 
# Manifest present and valid = correct (don't report)

import base64
import enum
import json
import os

import requests

schema_v1 = "https://keyfactor.github.io/integration-manifest-schema.json"
schema_v2 = "https://keyfactor.github.io/v2/integration-manifest-schema.json"

class RepoState(enum.Enum):
    NO_MANIFEST = 1
    OLDER_VERSION = 2
    INVALID = 3
    VALID = 4

class RepoResult:
    def __init__(self, owner, repo, state, details=None):
        self.owner = owner
        self.repo = repo
        self.state = state
        self.details = details

def ClassifyRepo(owner, repo) -> RepoResult:
    manifest_content = get_file_content(owner, repo, "integration-manifest.json")
    if manifest_content:
        try:
            manifest = json.loads(manifest_content)
            # Get the schema version
            schema_version = manifest.get("$schema", None)
            if schema_version == schema_v1:
                return RepoResult(owner, repo, RepoState.OLDER_VERSION)
            elif schema_version == schema_v2:
                # Validate the schema
                # TODO: implement
                return RepoResult(owner, repo, RepoState.VALID)
            elif schema_version is None:
                return RepoResult(owner, repo, RepoState.INVALID, details="Unknown or missing schema version")
            else:
                return RepoResult(owner, repo, RepoState.INVALID, details=f"Unrecognized schema version: {schema_version}")
        except json.JSONDecodeError as e:
            return RepoResult(owner, repo, RepoState.INVALID, details=f"JSON decode error: {e}")
        except Exception as e:
            return RepoResult(owner, repo, RepoState.INVALID, details=f"Error processing manifest: {e}")
    else:
        return RepoResult(owner, repo, RepoState.NO_MANIFEST)

HEADERS = {
    "Authorization": f"token {os.getenv('GITHUB_TOKEN')}", 
    "Accept": "application/vnd.github.raw+json"
}
GITHUB_API = "https://api.github.com"

def get_file_content(owner, repo, path):
    url = f"{GITHUB_API}/repos/{owner}/{repo}/contents/{path}"
    resp = requests.get(url, headers=HEADERS)
    if resp.status_code == 200:
        content = resp.content
        return content.decode('utf-8')
    else:
        return None

