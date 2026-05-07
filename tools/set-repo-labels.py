#!/usr/bin/env uv run
#/// script
# requirements: ["requests", "tqdm"]
# ///

import os
import sys
import requests
import json
from tqdm import tqdm

GITHUB_API = "https://api.github.com"
ORG = "keyfactor"
TOKEN = os.environ.get("GITHUB_TOKEN")

if not TOKEN:
	print("GITHUB_TOKEN environment variable not set.")
	sys.exit(1)

HEADERS = {
	"Authorization": f"token {TOKEN}",
	"Accept": "application/vnd.github+json"
}

def get_repos(org):
	repos = []
	page = 1
	while True:
		url = f"{GITHUB_API}/orgs/{org}/repos?per_page=100&page={page}"
		resp = requests.get(url, headers=HEADERS)
		if resp.status_code != 200:
			print(f"Failed to get repos: {resp.status_code}")
			break
		data = resp.json()
		if not data:
			break
		repos.extend(data)
		page += 1
	return repos

def get_file_content(owner, repo, path):
	url = f"{GITHUB_API}/repos/{owner}/{repo}/contents/{path}"
	resp = requests.get(url, headers=HEADERS)
	if resp.status_code == 200:
		content = resp.json().get('content')
		if content:
			import base64
			return base64.b64decode(content).decode('utf-8')
	return None

def getRepoProperties(owner, repo):
    # Use the custom properties API
    # See: https://docs.github.com/en/rest/repos/custom-properties?apiVersion=2022-11-28
    properties_url = f"{GITHUB_API}/repos/{owner}/{repo}/properties/values"
    resp = requests.get(properties_url, headers=HEADERS)
    if resp.status_code == 200:
		# we get pairs of { "name": "IsIntegration", "value": true } etc
        properties = resp.json().get('properties', [])
        return {prop['property-name']: prop['value'] for prop in properties}
    else:
        #tqdm.write(f"Failed to get custom properties for {repo}: {resp.status_code} {resp.text}")
        return {}
def SetRepoProperties(owner, repo, properties):
    # Use the custom properties API
    # See: https://docs.github.com/en/rest/repos/custom-properties?apiVersion=2022-11-28
	
    # GH expects a list of { "property-name": "IsIntegration", "value": true } etc, so we need to transform our dict into that format
    properties_list = [{"property_name": key, "value": value} for key, value in properties.items()]

    properties_url = f"{GITHUB_API}/repos/{owner}/{repo}/properties/values"
    # The schema must be defined in the repo or org. We'll assume the schema allows these fields.
    resp = requests.patch(properties_url, headers=HEADERS, json={
		"properties": properties_list,
		"owner": owner,
        "repo": repo
    })
    if resp.status_code not in (200, 204):
        tqdm.write(f"Failed to set custom properties for {repo}: {resp.status_code} {resp.text}")

def set_repo_properties(owner, repo, is_integration, integration_type=None):
    properties = {
        "IsIntegration": "true" if is_integration else "false",
        "IntegrationType": integration_type if integration_type else None
    }
    SetRepoProperties(owner, repo, properties)

def main():
	repos = get_repos(ORG)
	tqdm.write(f"Found {len(repos)} repositories in organization '{ORG}'")
	for repo in tqdm(repos, desc="Processing repositories"):
		repo_name = repo['name']
		manifest_content = get_file_content(ORG, repo_name, 'integration-manifest.json')
		if manifest_content:
			try:
				manifest = json.loads(manifest_content)
				integration_type = manifest.get('integration_type')
				if integration_type:
					set_repo_properties(ORG, repo_name, True, integration_type)
					tqdm.write(f"{repo_name}: Set IsIntegration=True, IntegrationType={integration_type}")
				else:
					#set_repo_properties(ORG, repo_name, True)
					tqdm.write(f"{repo_name}: Has integration manifest but no integration_type specified")
			except Exception as e:
				tqdm.write(f"{repo_name}: Failed to parse manifest: {e}")
		else:
			pass 
			#tqdm.write(f"{repo_name}: No integration-manifest.json found")

if __name__ == "__main__":
	main()
